use super::ast;
use super::vil;

pub fn generate(ast: &ast::Program) -> Result<vil::Program, String> {
    let mut generator = Generator {
        program: vil::Program {
            declarations: Vec::new(),
        },
        label_counter: 0,
        register_counter: 0,
        symbol_counter: 0,
    };
    generator.generate_program(ast);
    Ok(generator.program)
}

struct Generator {
    program: vil::Program,
    label_counter: u32,
    register_counter: u32,
    symbol_counter: u32,
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
            }
        }
    }

    fn generate_function_declaration(&mut self, declaration: &ast::FunctionDeclaration) {
        let vil_declaration = vil::FunctionDeclaration {
            name: declaration.name.clone(),
            // TODO
            parameters: Vec::new(),
            // TODO
            return_type: vil::Type::I64,
            blocks: Vec::new(),
        };
        self.program.declarations.push(vil_declaration);
        let label = self.claim_label(&declaration.name);
        self.start_block(label);

        self.generate_block(&declaration.body);
    }

    fn generate_expression(&mut self, expr: &ast::Expression) -> vil::Register {
        let r = self.claim_register();
        match &expr.kind {
            ast::ExpressionKind::Boolean(x) => {
                self.push(vil::Instruction::Set(r.clone(), vil::Immediate(*x as i64)));
            }
            ast::ExpressionKind::Integer(x) => {
                self.push(vil::Instruction::Set(r.clone(), vil::Immediate(*x)));
            }
            ast::ExpressionKind::Binary(b) => self.generate_binary_expression(&b, r.clone()),
            _ => {
                // TODO
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
                self.push(vil::Instruction::CmpEq(left, right));
                self.push(vil::Instruction::SetCmp(r));
            }
            ast::BinaryOpType::GreaterThan => {
                self.push(vil::Instruction::CmpGt(left, right));
                self.push(vil::Instruction::SetCmp(r));
            }
            ast::BinaryOpType::GreaterThanEquals => {
                self.push(vil::Instruction::CmpGte(left, right));
                self.push(vil::Instruction::SetCmp(r));
            }
            ast::BinaryOpType::LessThan => {
                self.push(vil::Instruction::CmpLt(left, right));
                self.push(vil::Instruction::SetCmp(r));
            }
            ast::BinaryOpType::LessThanEquals => {
                self.push(vil::Instruction::CmpLte(left, right));
                self.push(vil::Instruction::SetCmp(r));
            }
            ast::BinaryOpType::Multiply => {
                self.push(vil::Instruction::Mul(r, left, right));
            }
            ast::BinaryOpType::NotEquals => {
                self.push(vil::Instruction::CmpNeq(left, right));
                self.push(vil::Instruction::SetCmp(r));
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

    fn generate_block(&mut self, block: &Vec<ast::Statement>) {
        for stmt in block {
            self.generate_statement(stmt);
        }
    }

    fn generate_statement(&mut self, stmt: &ast::Statement) {
        match stmt {
            ast::Statement::Assign(stmt) => self.generate_assign_statement(stmt),
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

    fn generate_let_statement(&mut self, stmt: &ast::LetStatement) {
        let entry = stmt.symbol.entry.as_ref().unwrap();
        let symbol = vil::Memory(entry.unique_name.clone());
        self.push(vil::Instruction::Alloca(symbol.clone(), 8));
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
        //   br loop, loop_end
        //
        // loop:
        //   <body>
        //   br loop_cond
        //
        // loop_end:
        let cond_label = self.claim_label("while_cond");
        let loop_label = self.claim_label("while");
        let end_label = self.claim_label("while_end");

        self.start_block(cond_label.clone());
        let register = self.generate_expression(&stmt.condition);
        let tmp = self.claim_register();
        self.push(vil::Instruction::Set(tmp.clone(), vil::Immediate(0)));
        self.push(vil::Instruction::CmpEq(register, tmp));
        self.set_exit(vil::ExitInstruction::JumpIf(
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
