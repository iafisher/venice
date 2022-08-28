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
        label_counter: 0,
        register_counter: 0,
        string_counter: 0,
    };
    generator.generate_program(ast);
    Ok(generator.program)
}

struct Generator {
    // The program which is incrementally built up.
    program: vil::Program,
    info: Option<ast::FunctionInfo>,

    // Counters for generating unique symbols.
    label_counter: u32,
    string_counter: u32,

    // The counter of current registers in use.
    register_counter: u8,
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
        let vil_declaration = vil::FunctionDeclaration {
            name: name.clone(),
            blocks: Vec::new(),
        };
        self.info = Some(declaration.info.clone());
        self.program.declarations.push(vil_declaration);
        let label = self.claim_label(name);
        self.start_block(label, None);

        // Save callee-save registers.
        for i in 2..vil::GP_REGISTER_COUNT {
            self.push(vil::InstructionKind::CalleeSave(vil::Register::gp(i)));
        }

        let stack_frame_size = self.info.as_ref().unwrap().stack_frame_size;
        self.push(vil::InstructionKind::FrameSetUp(stack_frame_size));

        // Move parameters from registers onto the stack.
        for (i, parameter) in declaration.parameters.iter().enumerate() {
            self.push_with_comment(
                vil::InstructionKind::Store(
                    vil::Register::param(i as u8),
                    parameter.name.stack_offset,
                ),
                &parameter.name.unique_name,
            );
        }

        self.generate_block(&declaration.body);
    }

    fn generate_expression(&mut self, expr: &ast::Expression) -> vil::Register {
        let r = self.claim_register();

        use ast::ExpressionKind::*;
        match &expr.kind {
            Boolean(x) => {
                self.push(vil::InstructionKind::Set(
                    r,
                    vil::Immediate::Integer(*x as i64),
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
            Symbol(symbol) => {
                self.push_with_comment(
                    vil::InstructionKind::Load(r, symbol.stack_offset),
                    &symbol.unique_name,
                );
            }
            x => {
                panic!(
                    "internal error: expression type not implemented: {:?}",
                    expr
                );
            }
        }
        r
    }

    // A variant of `generate_expression` that generates more efficient code when an expression is
    // being used as a control-flow condition, e.g. in an `if` statement or a `while` loop.
    fn generate_expression_as_condition(
        &mut self,
        expr: &ast::Expression,
        true_label: vil::Label,
        false_label: vil::Label,
    ) {
        if let ast::ExpressionKind::Comparison(cmp_expr) = &expr.kind {
            let left = self.generate_expression(&cmp_expr.left);
            let right = self.generate_expression(&cmp_expr.right);

            self.push(vil::InstructionKind::Cmp(left, right));
            let exit = get_comparison_instruction(cmp_expr.op, true_label, false_label);
            self.push(exit);
        } else {
            let register = self.generate_expression(expr);
            let tmp = self.claim_register();
            self.push(vil::InstructionKind::Set(tmp, vil::Immediate::Integer(1)));
            self.push(vil::InstructionKind::Cmp(register, tmp));
            self.push(vil::InstructionKind::JumpEq(true_label, false_label));
        }
    }

    fn generate_binary_expression(&mut self, expr: &ast::BinaryExpression, r: vil::Register) {
        let old_register_counter = r.index() + 1;

        let left = self.generate_expression(&expr.left);
        let right = self.generate_expression(&expr.right);

        use common::BinaryOpType::*;
        match expr.op {
            Add => {
                self.push(vil::InstructionKind::Add(r, left, right));
            }
            Divide => {
                self.push(vil::InstructionKind::Div(r, left, right));
            }
            Multiply => {
                self.push(vil::InstructionKind::Mul(r, left, right));
            }
            Subtract => {
                self.push(vil::InstructionKind::Sub(r, left, right));
            }
            _ => {
                panic!("internal error: operator not implemented: {:?}", expr.op);
            }
        }

        self.reset_register_counter(old_register_counter);
    }

    fn generate_unary_expression(&mut self, expr: &ast::UnaryExpression, r: vil::Register) {
        let operand = self.generate_expression(&expr.operand);

        use common::UnaryOpType::*;
        match expr.op {
            Negate => {
                self.push(vil::InstructionKind::Negate(r, operand));
            }
            Not => {
                self.push(vil::InstructionKind::LogicalNot(r, operand));
            }
        }
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

        for (i, argument) in expr.arguments.iter().enumerate() {
            let argument_register = self.generate_expression(argument);
            self.push(vil::InstructionKind::Move(
                vil::Register::param(i.try_into().unwrap()),
                argument_register,
            ));
        }

        let unique_name = &expr.function.unique_name;
        if expr.function.external {
            self.program.externs.push(unique_name.clone());
        }

        // Save caller-save registers.
        self.push(vil::InstructionKind::CallerSave(vil::Register::gp(0)));
        self.push(vil::InstructionKind::CallerSave(vil::Register::gp(1)));

        if expr.variadic {
            self.push(vil::InstructionKind::CallVariadic(vil::FunctionLabel(
                unique_name.clone(),
            )));
        } else {
            self.push(vil::InstructionKind::Call(vil::FunctionLabel(
                unique_name.clone(),
            )));
        }

        // Restore caller-save registers.
        self.push(vil::InstructionKind::CallerRestore(vil::Register::gp(1)));
        self.push(vil::InstructionKind::CallerRestore(vil::Register::gp(0)));

        self.push(vil::InstructionKind::Move(r, vil::Register::Return));
    }

    fn generate_block(&mut self, block: &[ast::Statement]) {
        for stmt in block {
            self.generate_statement(stmt);
            // Reset register counter in between statements. Any value that a statement
            // produces that must persist (e.g., `let` bindings) must be stored in
            // memory.
            self.reset_register_counter(0);
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
        self.push(vil::InstructionKind::Move(vil::Register::Return, register));
        self.push(vil::InstructionKind::FrameTearDown(
            self.info.as_ref().unwrap().stack_frame_size,
        ));

        // Restore callee-save registers.
        for i in (2..vil::GP_REGISTER_COUNT).rev() {
            self.push(vil::InstructionKind::CalleeRestore(vil::Register::gp(i)));
        }

        self.push(vil::InstructionKind::Ret);
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

    fn claim_register(&mut self) -> vil::Register {
        let register = vil::Register::gp(self.register_counter);
        self.register_counter += 1;
        register
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

    fn reset_register_counter(&mut self, count: u8) {
        self.register_counter = count;
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
        Equals => vil::InstructionKind::JumpEq(true_label, false_label),
        GreaterThan => vil::InstructionKind::JumpGt(true_label, false_label),
        GreaterThanEquals => vil::InstructionKind::JumpGte(true_label, false_label),
        LessThan => vil::InstructionKind::JumpLt(true_label, false_label),
        LessThanEquals => vil::InstructionKind::JumpLte(true_label, false_label),
        NotEquals => vil::InstructionKind::JumpNeq(true_label, false_label),
    }
}
