// Code generation from an abstract syntax tree to a VIL program.

use std::collections::BTreeMap;

use super::ast;
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
        label_counter: 0,
        register_counter: 0,
        symbol_counter: 0,
        string_counter: 0,
    };
    generator.generate_program(ast);
    Ok(generator.program)
}

struct Generator {
    // The program which is incrementally built up.
    program: vil::Program,

    // Counters for generating unique symbols.
    label_counter: u32,
    symbol_counter: u32,
    string_counter: u32,

    // The counter of current registers in use.
    register_counter: u32,
}

impl Generator {
    fn generate_program(&mut self, program: &ast::Program) {
        for declaration in &program.declarations {
            self.generate_declaration(declaration);
        }
    }

    fn generate_declaration(&mut self, declaration: &ast::Declaration) {
        match declaration {
            ast::Declaration::Function(decl) => self.generate_function_declaration(decl),
            _ => {
                // TODO
                self.push(vil::Instruction::ToDo(String::from("declaration")));
            }
        }
    }

    fn generate_function_declaration(&mut self, declaration: &ast::FunctionDeclaration) {
        let name = &declaration.name.entry.as_ref().unwrap().unique_name;
        let mut parameters = Vec::new();
        for parameter in &declaration.parameters {
            let entry = parameter.name.entry.as_ref().unwrap();
            parameters.push(vil::FunctionParameter {
                name: parameter.name.name.clone(),
                unique_name: entry.unique_name.clone(),
                // TODO
                type_: vil::Type::I64,
            });
        }

        let stack_frame_size = 8 * (parameters.len() as u32);
        let vil_declaration = vil::FunctionDeclaration {
            name: name.clone(),
            parameters: parameters,
            // TODO
            return_type: vil::Type::I64,
            blocks: Vec::new(),
            // TODO: parameters are not necessarily all 8 bytes
            stack_frame_size: stack_frame_size,
        };
        self.program.declarations.push(vil_declaration);
        let label = self.claim_label(&name);
        self.start_block(label);

        self.generate_block(&declaration.body);
    }

    fn generate_expression(&mut self, expr: &ast::Expression) -> vil::Register {
        let r = self.claim_register();
        match &expr.kind {
            ast::ExpressionKind::Boolean(x) => {
                self.push(vil::Instruction::Set(
                    r.clone(),
                    vil::Immediate::Integer(*x as i64),
                ));
            }
            ast::ExpressionKind::Integer(x) => {
                self.push(vil::Instruction::Set(
                    r.clone(),
                    vil::Immediate::Integer(*x),
                ));
            }
            ast::ExpressionKind::String(s) => {
                let label = self.claim_string_label();
                self.program.strings.insert(label.clone(), s.clone());
                self.push(vil::Instruction::Set(
                    r.clone(),
                    vil::Immediate::Label(label),
                ));
            }
            ast::ExpressionKind::Binary(b) => self.generate_binary_expression(&b, r.clone()),
            ast::ExpressionKind::Call(e) => self.generate_call_expression(e, r.clone()),
            ast::ExpressionKind::Symbol(symbol) => {
                let entry = symbol.entry.as_ref().unwrap();
                self.push(vil::Instruction::Load(
                    r.clone(),
                    vil::Memory(entry.unique_name.clone()),
                    0,
                ));
            }
            x => {
                // TODO
                self.push(vil::Instruction::ToDo(format!("{:?}", x)));
            }
        }
        r
    }

    fn generate_binary_expression(&mut self, expr: &ast::BinaryExpression, r: vil::Register) {
        let old_register_counter = r.0 + 1;

        let left = self.generate_expression(&expr.left);
        let right = self.generate_expression(&expr.right);

        match expr.op {
            ast::BinaryOpType::Add => {
                self.push(vil::Instruction::Add(r, left, right));
            }
            ast::BinaryOpType::Divide => {
                self.push(vil::Instruction::Div(r, left, right));
            }
            ast::BinaryOpType::Equals => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::GreaterThan => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::GreaterThanEquals => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::LessThan => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::LessThanEquals => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::Multiply => {
                self.push(vil::Instruction::Mul(r, left, right));
            }
            ast::BinaryOpType::NotEquals => {
                self.generate_binary_comparison(expr.op, left, right, r);
            }
            ast::BinaryOpType::Subtract => {
                self.push(vil::Instruction::Sub(r, left, right));
            }
            _ => {
                // TODO
            }
        }

        self.register_counter = old_register_counter;
    }

    fn generate_binary_comparison(
        &mut self,
        op: ast::BinaryOpType,
        left: vil::Register,
        right: vil::Register,
        r: vil::Register,
    ) {
        let true_label = self.claim_label("eq");
        let false_label = self.claim_label("eq");
        let end_label = self.claim_label("eq_end");

        self.push(vil::Instruction::Cmp(left, right));

        let exit = match op {
            ast::BinaryOpType::Equals => {
                vil::ExitInstruction::JumpEq(true_label.clone(), false_label.clone())
            }
            ast::BinaryOpType::GreaterThan => {
                vil::ExitInstruction::JumpGt(true_label.clone(), false_label.clone())
            }
            ast::BinaryOpType::GreaterThanEquals => {
                vil::ExitInstruction::JumpGte(true_label.clone(), false_label.clone())
            }
            ast::BinaryOpType::LessThan => {
                vil::ExitInstruction::JumpLt(true_label.clone(), false_label.clone())
            }
            ast::BinaryOpType::LessThanEquals => {
                vil::ExitInstruction::JumpLte(true_label.clone(), false_label.clone())
            }
            ast::BinaryOpType::NotEquals => {
                vil::ExitInstruction::JumpNeq(true_label.clone(), false_label.clone())
            }
            _ => panic!("not a binary comparison"),
        };
        self.set_exit(exit);

        self.start_block(true_label);
        self.push(vil::Instruction::Set(r.clone(), vil::Immediate::Integer(1)));
        self.set_exit(vil::ExitInstruction::Jump(end_label.clone()));

        self.start_block(false_label);
        self.push(vil::Instruction::Set(r, vil::Immediate::Integer(0)));
        self.set_exit(vil::ExitInstruction::Jump(end_label.clone()));

        self.start_block(end_label);
    }

    fn generate_call_expression(&mut self, expr: &ast::CallExpression, r: vil::Register) {
        let mut argument_registers = Vec::new();
        for argument in &expr.arguments {
            let argument_register = self.generate_expression(argument);
            argument_registers.push(argument_register);
        }

        let entry = expr.function.entry.as_ref().unwrap();
        if entry.external {
            self.program.externs.push(entry.unique_name.clone());
        }

        self.push(vil::Instruction::Call(
            r,
            vil::FunctionLabel(entry.unique_name.clone()),
            argument_registers,
        ));
    }

    fn generate_block(&mut self, block: &Vec<ast::Statement>) {
        for stmt in block {
            self.generate_statement(stmt);
        }
    }

    fn generate_statement(&mut self, stmt: &ast::Statement) {
        match stmt {
            ast::Statement::Assign(stmt) => self.generate_assign_statement(stmt),
            ast::Statement::Expression(expr) => {
                let _ = self.generate_expression(expr);
            }
            ast::Statement::If(stmt) => self.generate_if_statement(stmt),
            ast::Statement::Let(stmt) => self.generate_let_statement(stmt),
            ast::Statement::Return(stmt) => self.generate_return_statement(stmt),
            ast::Statement::While(stmt) => self.generate_while_statement(stmt),
            _ => {
                // TODO
            }
        }
    }

    fn generate_assign_statement(&mut self, stmt: &ast::AssignStatement) {
        let entry = stmt.symbol.entry.as_ref().unwrap();
        let register = self.generate_expression(&stmt.value);
        self.push(vil::Instruction::Store(
            vil::Memory(entry.unique_name.clone()),
            register,
            0,
        ));
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

        let register = self.generate_expression(&stmt.if_clause.condition);
        let tmp = self.claim_register();
        self.push(vil::Instruction::Set(
            tmp.clone(),
            vil::Immediate::Integer(0),
        ));
        self.push(vil::Instruction::Cmp(register, tmp));
        self.set_exit(vil::ExitInstruction::JumpEq(
            false_label.clone(),
            true_label.clone(),
        ));

        self.start_block(true_label);
        self.generate_block(&stmt.if_clause.body);

        // TODO: must be a more elegant way to do this
        let mut needs_end_label = false;
        match self.current_block().exit {
            vil::ExitInstruction::Placeholder => {
                needs_end_label = true;
                self.set_exit(vil::ExitInstruction::Jump(end_label.clone()));
            }
            _ => {}
        }

        // TODO: handle elif_clauses

        self.start_block(false_label);
        self.generate_block(&stmt.else_clause);
        match self.current_block().exit {
            vil::ExitInstruction::Placeholder => {
                needs_end_label = true;
                self.set_exit(vil::ExitInstruction::Jump(end_label.clone()));
            }
            _ => {}
        }

        if needs_end_label {
            self.start_block(end_label);
        }
    }

    fn generate_let_statement(&mut self, stmt: &ast::LetStatement) {
        let entry = stmt.symbol.entry.as_ref().unwrap();
        let symbol = vil::Memory(entry.unique_name.clone());
        self.push(vil::Instruction::Alloca(symbol.clone(), 8));
        self.current_function().stack_frame_size += 8;

        let register = self.generate_expression(&stmt.value);
        self.push(vil::Instruction::Store(symbol, register, 0));
    }

    fn generate_return_statement(&mut self, stmt: &ast::ReturnStatement) {
        let register = self.generate_expression(&stmt.value);
        self.set_exit(vil::ExitInstruction::Ret(register));
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

        self.start_block(cond_label.clone());
        let register = self.generate_expression(&stmt.condition);
        let tmp = self.claim_register();
        self.push(vil::Instruction::Set(
            tmp.clone(),
            vil::Immediate::Integer(1),
        ));
        self.push(vil::Instruction::Cmp(register, tmp));
        self.set_exit(vil::ExitInstruction::JumpEq(
            loop_label.clone(),
            end_label.clone(),
        ));

        self.start_block(loop_label);
        self.generate_block(&stmt.body);
        self.set_exit(vil::ExitInstruction::Jump(cond_label));

        self.start_block(end_label);
    }

    fn start_block(&mut self, label: vil::Label) {
        let block = vil::Block {
            name: label.0.clone(),
            instructions: Vec::new(),
            exit: vil::ExitInstruction::Placeholder,
        };

        if self.current_function().blocks.len() > 0 {
            if let vil::ExitInstruction::Placeholder = self.current_block().exit {
                self.set_exit(vil::ExitInstruction::Jump(label));
            }
        }

        self.current_function().blocks.push(block);
    }

    fn claim_label(&mut self, prefix: &str) -> vil::Label {
        let label = format!("{}_{}", prefix, self.label_counter);
        self.label_counter += 1;
        vil::Label(label)
    }

    fn claim_symbol(&mut self, prefix: &str) -> vil::Memory {
        let symbol = format!("{}_{}", prefix, self.symbol_counter);
        self.symbol_counter += 1;
        vil::Memory(symbol)
    }

    fn claim_register(&mut self) -> vil::Register {
        let register = vil::Register(self.register_counter);
        self.register_counter += 1;
        register
    }

    fn claim_string_label(&mut self) -> String {
        let label = format!("s_{}", self.string_counter);
        self.string_counter += 1;
        label
    }

    fn set_exit(&mut self, exit: vil::ExitInstruction) {
        self.current_block().exit = exit;
    }

    fn push(&mut self, instruction: vil::Instruction) {
        self.current_block().instructions.push(instruction)
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
