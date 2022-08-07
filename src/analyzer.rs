use super::ast;
use super::errors;
use std::collections::HashMap;

fn analyze(ast: &mut ast::Program) -> Result<(), errors::VeniceError> {
    let mut analyzer = Analyzer::new();
    analyzer.analyze_program(ast)
}

struct SymbolTable {
    symbols: HashMap<String, SymbolTableEntry>,
}

#[derive(Clone)]
struct SymbolTableEntry {
    pub type_: ast::Type,
    pub constant: bool,
}

impl SymbolTable {
    pub fn new() -> Self {
        SymbolTable {
            symbols: HashMap::new(),
        }
    }

    pub fn builtin_types() -> Self {
        let mut symbols = HashMap::new();
        symbols.insert(
            String::from("i64"),
            SymbolTableEntry {
                type_: ast::Type::I64,
                constant: true,
            },
        );
        symbols.insert(
            String::from("bool"),
            SymbolTableEntry {
                type_: ast::Type::Boolean,
                constant: true,
            },
        );
        symbols.insert(
            String::from("string"),
            SymbolTableEntry {
                type_: ast::Type::Str,
                constant: true,
            },
        );

        SymbolTable { symbols: symbols }
    }

    pub fn get(&self, key: &str) -> Option<SymbolTableEntry> {
        self.symbols.get(key).cloned()
    }

    pub fn insert(&mut self, key: &str, entry: SymbolTableEntry) {
        self.symbols.insert(String::from(key), entry.clone());
    }

    pub fn remove(&mut self, key: &str) {
        self.symbols.remove(key);
    }
}

struct Analyzer {
    symbols: SymbolTable,
    types: SymbolTable,
    errors: Vec<errors::VeniceError>,
}

impl Analyzer {
    fn new() -> Self {
        Analyzer {
            symbols: SymbolTable::new(),
            types: SymbolTable::builtin_types(),
            errors: Vec::new(),
        }
    }

    fn analyze_program(&mut self, ast: &mut ast::Program) -> Result<(), errors::VeniceError> {
        for mut declaration in &mut ast.declarations {
            self.analyze_declaration(&mut declaration)
        }

        if self.errors.len() > 0 {
            // TODO: return all errors, not just the first one.
            Err(self.errors[0].clone())
        } else {
            Ok(())
        }
    }

    fn analyze_declaration(&mut self, declaration: &mut ast::Declaration) {
        match declaration {
            ast::Declaration::Function(d) => self.analyze_function_declaration(d),
            ast::Declaration::Const(d) => self.analyze_const_declaration(d),
            ast::Declaration::Record(d) => self.analyze_record_declaration(d),
        }
    }

    fn analyze_function_declaration(&mut self, declaration: &mut ast::FunctionDeclaration) {
        declaration.semantic_return_type = self.resolve_type(&declaration.return_type);
        let mut parameter_types = Vec::new();
        for parameter in &mut declaration.parameters {
            let t = self.resolve_type(&parameter.type_);
            parameter.semantic_type = t.clone();
            self.symbols.insert(
                &parameter.name,
                SymbolTableEntry {
                    type_: t.clone(),
                    constant: false,
                },
            );
            parameter_types.push(t);
        }
        self.types.insert(
            &declaration.name,
            SymbolTableEntry {
                type_: ast::Type::Function {
                    parameters: parameter_types,
                    return_type: Box::new(declaration.semantic_return_type.clone()),
                },
                constant: true,
            },
        );

        for stmt in &declaration.body {
            self.analyze_statement(stmt);
        }

        // TODO: better approach would be to have a separate symbol table for the
        // function's scope.
        for parameter in &mut declaration.parameters {
            self.symbols.remove(&parameter.name);
        }
    }

    fn analyze_const_declaration(&mut self, declaration: &mut ast::ConstDeclaration) {
        self.analyze_expression(&declaration.value);
        declaration.semantic_type = self.resolve_type(&declaration.type_);
        if !declaration
            .semantic_type
            .matches(&declaration.value.semantic_type)
        {
            let msg = format!(
                "expected {}, got {}",
                declaration.semantic_type, declaration.value.semantic_type,
            );
            self.error(&msg);
        }

        self.symbols.insert(
            &declaration.symbol,
            SymbolTableEntry {
                type_: declaration.semantic_type.clone(),
                constant: true,
            },
        );
    }

    fn analyze_record_declaration(&mut self, declaration: &mut ast::RecordDeclaration) {
        // TODO
    }

    fn analyze_statement(&self, stmt: &ast::Statement) {
        // TODO
    }

    fn analyze_expression(&self, expr: &ast::Expression) {
        // TODO
    }

    fn resolve_type(&self, type_: &ast::SyntacticType) -> ast::Type {
        ast::Type::Error
    }

    fn error(&mut self, message: &str) {
        self.errors.push(errors::VeniceError::new(message));
    }
}
