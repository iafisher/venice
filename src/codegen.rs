// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
// Code generation from an abstract syntax tree to a VIL program.

use std::collections::BTreeMap;

use super::ast;
use super::common;
use super::errors;
use super::vil;

/// Generates a VIL program from an abstract syntax tree.
pub fn generate(ast: &ast::Program) -> Result<vil::Program, errors::VeniceError> {
    let mut generator = Generator {
        program: vil::Program {
            externs: Vec::new(),
            declarations: Vec::new(),
            strings: BTreeMap::new(),
        },
        info: None,
        return_label: vil::Label(String::new()),
        label_counter: 0,
        string_counter: 0,
    };
    generator.generate_program(ast);
    Ok(generator.program)
}

struct Generator {
    // The program which is incrementally built up.
    program: vil::Program,
    info: Option<ast::FunctionInfo>,
    return_label: vil::Label,

    // Counters for generating unique symbols.
    label_counter: u32,
    string_counter: u32,
}

impl Generator {
    fn generate_program(&mut self, program: &ast::Program) {
        for declaration in &program.declarations {
            self.generate_declaration(declaration);
        }
    }

    fn generate_declaration(&mut self, declaration: &ast::Declaration) {
        use ast::Declaration::*;
        match declaration {
            Function(decl) => self.generate_function_declaration(decl),
            _ => {
                panic!("internal error: record and const declarations are not yet supported");
            }
        }
    }

    fn generate_function_declaration(&mut self, declaration: &ast::FunctionDeclaration) {
        let name = &declaration.name.unique_name;
        self.info = Some(declaration.info.clone());

        let mut parameters = Vec::new();
        for parameter in &declaration.parameters {
            parameters.push(vil::FunctionParameter {
                stack_offset: parameter.name.stack_offset,
            });
        }

        let vil_declaration = vil::FunctionDeclaration {
            name: name.clone(),
            parameters,
            blocks: Vec::new(),
            stack_frame_size: self.info.as_ref().unwrap().stack_frame_size,
        };

        self.program.declarations.push(vil_declaration);
        let label = self.claim_label(name);
        self.return_label = self.claim_label(&format!("{}_return", name));

        self.start_block(label, None);
        self.generate_block(&declaration.body);
        self.start_block(self.return_label.clone(), None);
    }

    fn generate_expression(&mut self, expr: &ast::Expression) -> vil::Register {
        let r = vil::Register::new(expr.max_register_needed);

        use ast::ExpressionKind::*;
        match &expr.kind {
            Boolean(x) => {
                self.push(vil::InstructionKind::Set(
                    r,
                    vil::Immediate::Integer(i64::try_from(*x).unwrap()),
                ));
            }
            Integer(x) => {
                self.push(vil::InstructionKind::Set(r, vil::Immediate::Integer(*x)));
            }
            String(s) => {
                let label = self.claim_string_label();
                self.program.strings.insert(label.clone(), s.clone());
                self.push(vil::InstructionKind::Set(r, vil::Immediate::Label(label)));
            }
            Binary(b) => self.generate_binary_expression(b, r),
            Unary(e) => self.generate_unary_expression(e, r),
            Comparison(b) => self.generate_comparison_expression(b, r),
            Call(e) => self.generate_call_expression(e, r),
            If(e) => self.generate_if_expression(e, r),
            Symbol(symbol) => {
                self.push_with_comment(
                    vil::InstructionKind::Load(r, symbol.stack_offset),
                    &symbol.unique_name,
                );
            }
            _x => {
                panic!(
                    "internal error: expression type not implemented: {:?}",
                    expr
                );
            }
        }
        r
    }

    /// A variant of `generate_expression` that generates more efficient code when an expression is
    /// being used as a control-flow condition, e.g. in an `if` statement or a `while` loop.
    fn generate_expression_as_condition(
        &mut self,
        expr: &ast::Expression,
        true_label: vil::Label,
        false_label: vil::Label,
    ) {
        let r = vil::Register::new(expr.max_register_needed);
        if let ast::ExpressionKind::Comparison(cmp_expr) = &expr.kind {
            let (left, right) =
                self.generate_generic_binary_expression(&cmp_expr.left, &cmp_expr.right, r);
            self.push(vil::InstructionKind::Cmp(left, right));
            let exit = get_comparison_instruction(cmp_expr.op, true_label, false_label);
            self.push(exit);
        } else {
            let register = self.generate_expression(expr);
            let scratch = vil::Register::scratch();
            self.push(vil::InstructionKind::Set(
                scratch,
                vil::Immediate::Integer(1),
            ));
            self.push(vil::InstructionKind::Cmp(register, scratch));
            self.push(vil::InstructionKind::JumpIf(
                vil::JumpCondition::Eq,
                true_label,
                false_label,
            ));
        }
    }

    fn generate_binary_expression(&mut self, expr: &ast::BinaryExpression, r: vil::Register) {
        let (left, right) = self.generate_generic_binary_expression(&expr.left, &expr.right, r);

        use common::BinaryOpType::*;
        let op = match expr.op {
            Add => vil::BinaryOp::Add,
            Divide => vil::BinaryOp::Div,
            Multiply => vil::BinaryOp::Mul,
            Subtract => vil::BinaryOp::Sub,
            And | Or => {
                panic!(
                    "internal error: and/or expressions should have been converted by the analyzer"
                );
            }
            _ => {
                panic!("internal error: operator not implemented: {:?}", expr.op);
            }
        };
        self.push(vil::InstructionKind::Binary(op, r, left, right));
    }

    /// Given a left and right expression and a target register, generates the code for the two
    /// expressions and returns (left, right), the pair of registers that the results will be
    /// placed in.
    ///
    /// The target register is necessary in the case where both expressions use the same number of
    /// registers and an additional register is needed to store one of the results.
    fn generate_generic_binary_expression(
        &mut self,
        left: &ast::Expression,
        right: &ast::Expression,
        r: vil::Register,
    ) -> (vil::Register, vil::Register) {
        if left.max_register_needed < right.max_register_needed {
            let register_right = self.generate_expression(right);
            let register_left = self.generate_expression(left);
            (register_left, register_right)
        } else if left.max_register_needed == right.max_register_needed {
            let register_left = self.generate_expression(left);
            self.push(vil::InstructionKind::Move(r, register_left));
            let register_right = self.generate_expression(right);
            (r, register_right)
        } else {
            let register_left = self.generate_expression(left);
            let register_right = self.generate_expression(right);
            (register_left, register_right)
        }
    }

    fn generate_unary_expression(&mut self, expr: &ast::UnaryExpression, r: vil::Register) {
        let operand = self.generate_expression(&expr.operand);

        use common::UnaryOpType::*;
        let op = match expr.op {
            Negate => vil::UnaryOp::Negate,
            Not => vil::UnaryOp::LogicalNot,
        };
        self.push(vil::InstructionKind::Unary(op, r, operand));
    }

    fn generate_comparison_expression(
        &mut self,
        expr: &ast::ComparisonExpression,
        r: vil::Register,
    ) {
        let left = self.generate_expression(&expr.left);
        let right = self.generate_expression(&expr.right);

        let true_label = self.claim_label("eq");
        let false_label = self.claim_label("eq");
        let end_label = self.claim_label("eq_end");

        self.push(vil::InstructionKind::Cmp(left, right));

        let exit = get_comparison_instruction(expr.op, true_label.clone(), false_label.clone());
        self.start_block(true_label, Some(exit));
        self.push(vil::InstructionKind::Set(r, vil::Immediate::Integer(1)));

        self.start_block(
            false_label,
            Some(vil::InstructionKind::Jump(end_label.clone())),
        );
        self.push(vil::InstructionKind::Set(r, vil::Immediate::Integer(0)));

        self.start_block(
            end_label.clone(),
            Some(vil::InstructionKind::Jump(end_label)),
        );
    }

    fn generate_call_expression(&mut self, expr: &ast::CallExpression, r: vil::Register) {
        if expr.arguments.len() > 6 {
            panic!("internal error: compiler cannot handle more than 6 arguments")
        }

        let mut offsets = Vec::new();
        for (i, argument) in expr.arguments.iter().enumerate() {
            let argument_register = self.generate_expression(argument);
            if argument.stack_offset == 0 {
                panic!(
                    "internal error: argument {} has invalid stack offset in call to {}",
                    i, expr.function.unique_name
                );
            }
            self.push(vil::InstructionKind::Store(
                argument_register,
                argument.stack_offset,
            ));
            offsets.push(argument.stack_offset);
        }

        let unique_name = &expr.function.unique_name;
        if expr.function.external {
            self.program.externs.push(unique_name.clone());
        }

        self.push(vil::InstructionKind::Call {
            label: vil::FunctionLabel(unique_name.clone()),
            offsets,
            variadic: expr.variadic,
        });

        self.push(vil::InstructionKind::Move(r, vil::Register::ret()));
    }

    fn generate_if_expression(&mut self, expr: &ast::IfExpression, r: vil::Register) {
        let true_label = self.claim_label("if_true");
        let false_label = self.claim_label("if_false");
        let end_label = self.claim_label("if_end");

        let condition = self.generate_expression(&expr.condition);
        let scratch = vil::Register::scratch();
        self.push(vil::InstructionKind::Set(
            scratch,
            vil::Immediate::Integer(1),
        ));
        self.push(vil::InstructionKind::Cmp(condition, scratch));

        self.start_block(
            true_label.clone(),
            Some(vil::InstructionKind::JumpIf(
                vil::JumpCondition::Eq,
                true_label,
                false_label.clone(),
            )),
        );
        let true_value = self.generate_expression(&expr.true_value);
        self.push(vil::InstructionKind::Move(r.clone(), true_value));

        self.start_block(
            false_label,
            Some(vil::InstructionKind::Jump(end_label.clone())),
        );
        let false_value = self.generate_expression(&expr.false_value);
        self.push(vil::InstructionKind::Move(r.clone(), false_value));

        self.start_block(
            end_label.clone(),
            Some(vil::InstructionKind::Jump(end_label)),
        );
    }

    fn generate_block(&mut self, block: &[ast::Statement]) {
        for stmt in block {
            self.generate_statement(stmt);
        }
    }

    fn generate_statement(&mut self, stmt: &ast::Statement) {
        use ast::Statement::*;
        match stmt {
            Assign(stmt) => self.generate_assign_statement(stmt),
            Expression(expr) => {
                let _ = self.generate_expression(expr);
            }
            If(stmt) => self.generate_if_statement(stmt),
            Let(stmt) => self.generate_let_statement(stmt),
            Return(stmt) => self.generate_return_statement(stmt),
            While(stmt) => self.generate_while_statement(stmt),
            _ => {
                panic!("internal error: statement type not implemented: {:?}", stmt);
            }
        }
    }

    fn generate_assign_statement(&mut self, stmt: &ast::AssignStatement) {
        let register = self.generate_expression(&stmt.value);
        self.push_with_comment(
            vil::InstructionKind::Store(register, stmt.symbol.stack_offset),
            &stmt.symbol.unique_name,
        );
    }

    fn generate_if_statement(&mut self, stmt: &ast::IfStatement) {
        // if cond {
        //   body1
        // } else {
        //   body2
        // }
        //
        // becomes
        //
        //   <cond>
        //   jump_eq body1, body2
        //
        // body1:
        //   <body1>
        //   jump end
        //
        // body2:
        //   <body2>
        //   jump end
        //
        // end:

        let true_label = self.claim_label("if_true");
        let false_label = self.claim_label("if_false");
        let end_label = self.claim_label("if_end");

        self.generate_expression_as_condition(
            &stmt.condition,
            true_label.clone(),
            false_label.clone(),
        );

        self.start_block(true_label, None);
        self.generate_block(&stmt.body);

        self.start_block(
            false_label,
            Some(vil::InstructionKind::Jump(end_label.clone())),
        );
        self.generate_block(&stmt.else_body);

        self.start_block(
            end_label.clone(),
            Some(vil::InstructionKind::Jump(end_label)),
        );
    }

    fn generate_let_statement(&mut self, stmt: &ast::LetStatement) {
        let register = self.generate_expression(&stmt.value);
        self.push_with_comment(
            vil::InstructionKind::Store(register, stmt.symbol.stack_offset),
            &stmt.symbol.unique_name,
        );
    }

    fn generate_return_statement(&mut self, stmt: &ast::ReturnStatement) {
        let register = self.generate_expression(&stmt.value);
        self.push(vil::InstructionKind::Move(vil::Register::ret(), register));
        self.push(vil::InstructionKind::Jump(self.return_label.clone()));
    }

    fn generate_while_statement(&mut self, stmt: &ast::WhileStatement) {
        // while cond {
        //   body
        // }
        //
        // becomes
        //
        // loop_cond:
        //   <cond>
        //   jump_eq loop, loop_end
        //
        // loop:
        //   <body>
        //   jump loop_cond
        //
        // loop_end:
        let cond_label = self.claim_label("while_cond");
        let loop_label = self.claim_label("while");
        let end_label = self.claim_label("while_end");

        self.start_block(
            cond_label.clone(),
            Some(vil::InstructionKind::Jump(cond_label.clone())),
        );
        self.generate_expression_as_condition(
            &stmt.condition,
            loop_label.clone(),
            end_label.clone(),
        );

        self.start_block(loop_label, None);
        self.generate_block(&stmt.body);

        self.start_block(end_label, Some(vil::InstructionKind::Jump(cond_label)));
    }

    fn start_block(&mut self, label: vil::Label, exit_previous: Option<vil::InstructionKind>) {
        let block = vil::Block {
            name: label.0,
            instructions: Vec::new(),
        };

        if let Some(exit_previous) = exit_previous {
            self.push(exit_previous);
        }

        self.current_function().blocks.push(block);
    }

    fn claim_label(&mut self, prefix: &str) -> vil::Label {
        let label = format!("{}_{}", prefix, self.label_counter);
        self.label_counter += 1;
        vil::Label(label)
    }

    fn claim_string_label(&mut self) -> String {
        let label = format!("s_{}", self.string_counter);
        self.string_counter += 1;
        label
    }

    fn push(&mut self, instruction: vil::InstructionKind) {
        self.current_block().instructions.push(vil::Instruction {
            kind: instruction,
            comment: String::new(),
        })
    }

    fn push_with_comment(&mut self, instruction: vil::InstructionKind, comment: &str) {
        self.current_block().instructions.push(vil::Instruction {
            kind: instruction,
            comment: String::from(comment),
        })
    }

    fn current_function(&mut self) -> &mut vil::FunctionDeclaration {
        let index = self.program.declarations.len() - 1;
        &mut self.program.declarations[index]
    }

    fn current_block(&mut self) -> &mut vil::Block {
        let function = self.current_function();
        let index = function.blocks.len() - 1;
        &mut function.blocks[index]
    }
}

fn get_comparison_instruction(
    op: common::ComparisonOpType,
    true_label: vil::Label,
    false_label: vil::Label,
) -> vil::InstructionKind {
    use common::ComparisonOpType::*;
    match op {
        Equals => vil::InstructionKind::JumpIf(vil::JumpCondition::Eq, true_label, false_label),
        GreaterThan => {
            vil::InstructionKind::JumpIf(vil::JumpCondition::Gt, true_label, false_label)
        }
        GreaterThanEquals => {
            vil::InstructionKind::JumpIf(vil::JumpCondition::Gte, true_label, false_label)
        }
        LessThan => vil::InstructionKind::JumpIf(vil::JumpCondition::Lt, true_label, false_label),
        LessThanEquals => {
            vil::InstructionKind::JumpIf(vil::JumpCondition::Lte, true_label, false_label)
        }
        NotEquals => vil::InstructionKind::JumpIf(vil::JumpCondition::Neq, true_label, false_label),
    }
}
