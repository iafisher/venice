use super::ast;
use super::vil;
use super::x86;

pub fn generate_ir(ast: &ast::Program) -> Result<vil::Program, String> {
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
        // TODO
        self.claim_register()
    }

    fn generate_block(&mut self, block: &Vec<ast::Statement>) {
        for stmt in block {
            self.generate_statement(stmt);
        }
    }

    fn generate_statement(&mut self, stmt: &ast::Statement) {
        match stmt {
            ast::Statement::Let(stmt) => self.generate_let_statement(stmt),
            ast::Statement::Return(stmt) => self.generate_return_statement(stmt),
            ast::Statement::While(stmt) => self.generate_while_statement(stmt),
            _ => {
                // TODO
            }
        }
    }

    fn generate_let_statement(&mut self, stmt: &ast::LetStatement) {
        let symbol = self.claim_symbol(&stmt.symbol);
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

fn generate_x86(vil: vil::Program) -> Result<x86::Program, String> {
    Err(String::from("not implemented"))
}
