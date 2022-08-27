// The analyzer traverses the program's abstract syntax tree and augments it with type
// information.

use super::ast;
use super::common;
use super::errors;
use std::collections::HashMap;

/// Analyzes the abstract syntax tree. The nodes of the tree are mutated to add type
/// and symbol information.
pub fn analyze(ast: &mut ast::Program) -> Result<(), Vec<errors::VeniceError>> {
    let mut analyzer = Analyzer::new();
    analyzer.analyze_program(ast);
    if analyzer.errors.len() > 0 {
        Err(analyzer.errors.clone())
    } else {
        Ok(())
    }
}

struct SymbolTable {
    symbols: HashMap<String, ast::SymbolEntry>,
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
            ast::SymbolEntry {
                unique_name: String::new(),
                type_: ast::Type::I64,
                constant: true,
                external: false,
            },
        );
        symbols.insert(
            String::from("bool"),
            ast::SymbolEntry {
                unique_name: String::new(),
                type_: ast::Type::Boolean,
                constant: true,
                external: false,
            },
        );
        symbols.insert(
            String::from("string"),
            ast::SymbolEntry {
                unique_name: String::new(),
                type_: ast::Type::String,
                constant: true,
                external: false,
            },
        );
        symbols.insert(
            String::from("void"),
            ast::SymbolEntry {
                unique_name: String::new(),
                type_: ast::Type::Void,
                constant: true,
                external: false,
            },
        );

        SymbolTable { symbols: symbols }
    }

    pub fn builtin_globals() -> Self {
        let mut symbols = HashMap::new();
        // TODO: unique names here could conflict with actual user symbols.
        symbols.insert(
            String::from("println"),
            ast::SymbolEntry {
                unique_name: String::from("venice_println"),
                type_: ast::Type::Function {
                    parameters: vec![ast::Type::String],
                    return_type: Box::new(ast::Type::Void),
                },
                constant: true,
                external: true,
            },
        );
        // TODO: remove printint once there's a better way to print integers.
        symbols.insert(
            String::from("printint"),
            ast::SymbolEntry {
                unique_name: String::from("venice_printint"),
                type_: ast::Type::Function {
                    parameters: vec![ast::Type::I64],
                    return_type: Box::new(ast::Type::Void),
                },
                constant: true,
                external: true,
            },
        );

        SymbolTable { symbols }
    }

    pub fn get(&self, key: &str) -> Option<ast::SymbolEntry> {
        self.symbols.get(key).cloned()
    }

    pub fn insert(&mut self, key: &str, entry: ast::SymbolEntry) {
        self.symbols.insert(String::from(key), entry.clone());
    }

    pub fn remove(&mut self, key: &str) {
        self.symbols.remove(key);
    }
}

struct Analyzer {
    symbols: SymbolTable,
    types: SymbolTable,
    current_function_return_type: Option<ast::Type>,
    errors: Vec<errors::VeniceError>,
    unique_name_counter: u64,
}

impl Analyzer {
    fn new() -> Self {
        Analyzer {
            symbols: SymbolTable::builtin_globals(),
            types: SymbolTable::builtin_types(),
            current_function_return_type: None,
            errors: Vec::new(),
            unique_name_counter: 0,
        }
    }

    fn analyze_program(&mut self, ast: &mut ast::Program) {
        for mut declaration in &mut ast.declarations {
            self.analyze_declaration(&mut declaration)
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
            let unique_name = self.claim_unique_name(&parameter.name.name);
            let entry = ast::SymbolEntry {
                unique_name: unique_name,
                type_: t.clone(),
                constant: false,
                external: false,
            };
            self.symbols.insert(&parameter.name.name, entry.clone());
            parameter_types.push(t);
            parameter.name.entry = Some(entry);
        }

        let unique_name = if declaration.name.name == "main" {
            // Keep main's name the same so that the linker can find it.
            String::from("main")
        } else {
            self.claim_unique_name(&declaration.name.name)
        };

        let entry = ast::SymbolEntry {
            unique_name: unique_name,
            type_: ast::Type::Function {
                parameters: parameter_types,
                return_type: Box::new(declaration.semantic_return_type.clone()),
            },
            constant: true,
            external: false,
        };

        self.symbols.insert(&declaration.name.name, entry.clone());

        declaration.name.entry = Some(entry);

        self.current_function_return_type = Some(declaration.semantic_return_type.clone());
        self.analyze_block(&mut declaration.body);
        self.current_function_return_type = None;

        // TODO: A better approach would be to have a separate symbol table for the
        // function's scope. And actually this doesn't work at all because the symbols
        // defined in a function's body will persist after the function is finished.
        for parameter in &mut declaration.parameters {
            self.symbols.remove(&parameter.name.name);
        }
    }

    fn analyze_const_declaration(&mut self, declaration: &mut ast::ConstDeclaration) {
        self.analyze_expression(&mut declaration.value);
        declaration.semantic_type = self.resolve_type(&declaration.type_);
        if !declaration
            .semantic_type
            .matches(&declaration.value.semantic_type)
        {
            self.error_type_mismatch(
                &declaration.semantic_type,
                &declaration.value.semantic_type,
                declaration.location.clone(),
            );
        }

        let unique_name = self.claim_unique_name(&declaration.symbol);
        self.symbols.insert(
            &declaration.symbol,
            ast::SymbolEntry {
                unique_name: unique_name,
                type_: declaration.semantic_type.clone(),
                constant: true,
                external: false,
            },
        );
    }

    fn analyze_record_declaration(&mut self, declaration: &mut ast::RecordDeclaration) {
        // TODO
    }

    fn analyze_block(&mut self, block: &mut Vec<ast::Statement>) {
        for stmt in block {
            self.analyze_statement(stmt);
        }
    }

    fn analyze_statement(&mut self, stmt: &mut ast::Statement) {
        match stmt {
            ast::Statement::Let(s) => self.analyze_let_statement(s),
            ast::Statement::Assign(s) => self.analyze_assign_statement(s),
            ast::Statement::If(s) => self.analyze_if_statement(s),
            ast::Statement::While(s) => self.analyze_while_statement(s),
            ast::Statement::For(s) => self.analyze_for_statement(s),
            ast::Statement::Return(s) => self.analyze_return_statement(s),
            ast::Statement::Assert(s) => self.analyze_assert_statement(s),
            ast::Statement::Expression(expr) => {
                let _ = self.analyze_expression(expr);
            }
        }
    }

    fn analyze_let_statement(&mut self, stmt: &mut ast::LetStatement) {
        self.analyze_expression(&mut stmt.value);
        stmt.semantic_type = self.resolve_type(&stmt.type_);
        if !stmt.semantic_type.matches(&stmt.value.semantic_type) {
            self.error_type_mismatch(
                &stmt.semantic_type,
                &stmt.value.semantic_type,
                stmt.location.clone(),
            );
        }

        let unique_name = self.claim_unique_name(&stmt.symbol.name);
        let entry = ast::SymbolEntry {
            unique_name: unique_name,
            type_: stmt.semantic_type.clone(),
            constant: false,
            external: false,
        };

        stmt.symbol.entry = Some(entry.clone());
        self.symbols.insert(&stmt.symbol.name, entry);
    }

    fn analyze_assign_statement(&mut self, stmt: &mut ast::AssignStatement) {
        self.analyze_expression(&mut stmt.value);
        if let Some(entry) = self.symbols.get(&stmt.symbol.name) {
            stmt.symbol.entry = Some(entry.clone());
            if !entry.type_.matches(&stmt.value.semantic_type) {
                self.error_type_mismatch(
                    &entry.type_,
                    &stmt.value.semantic_type,
                    stmt.location.clone(),
                );
            }
        } else {
            let msg = format!("assignment to unknown symbol {}", stmt.symbol);
            self.error(&msg, stmt.location.clone());
        }
    }

    fn analyze_if_statement(&mut self, stmt: &mut ast::IfStatement) {
        let if_clause_type = self.analyze_expression(&mut stmt.if_clause.condition);
        if !if_clause_type.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(
                &ast::Type::Boolean,
                &if_clause_type,
                stmt.if_clause.condition.location.clone(),
            );
        }
        self.analyze_block(&mut stmt.if_clause.body);

        for elif_clause in &mut stmt.elif_clauses {
            let elif_clause_type = self.analyze_expression(&mut elif_clause.condition);
            if !elif_clause_type.matches(&ast::Type::Boolean) {
                self.error_type_mismatch(
                    &ast::Type::Boolean,
                    &elif_clause_type,
                    elif_clause.condition.location.clone(),
                );
            }
            self.analyze_block(&mut elif_clause.body);
        }

        self.analyze_block(&mut stmt.else_clause);
    }

    fn analyze_while_statement(&mut self, stmt: &mut ast::WhileStatement) {
        let condition_type = self.analyze_expression(&mut stmt.condition);
        if !condition_type.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(
                &ast::Type::Boolean,
                &condition_type,
                stmt.condition.location.clone(),
            );
        }
        self.analyze_block(&mut stmt.body);
    }

    fn analyze_for_statement(&mut self, stmt: &mut ast::ForStatement) {}

    fn analyze_return_statement(&mut self, stmt: &mut ast::ReturnStatement) {
        let actual_return_type = self.analyze_expression(&mut stmt.value);
        // TODO: Can the clone here be avoided?
        if let Some(expected_return_type) = self.current_function_return_type.clone() {
            if !expected_return_type.matches(&actual_return_type) {
                self.error_type_mismatch(
                    &expected_return_type,
                    &actual_return_type,
                    stmt.location.clone(),
                );
            }
        } else {
            self.error(
                "return statement outside of function",
                stmt.location.clone(),
            );
        }
    }

    fn analyze_assert_statement(&mut self, stmt: &mut ast::AssertStatement) {
        let type_ = self.analyze_expression(&mut stmt.condition);
        if !type_.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(&ast::Type::Boolean, &type_, stmt.condition.location.clone());
        }
    }

    fn analyze_expression(&mut self, expr: &mut ast::Expression) -> ast::Type {
        expr.semantic_type = match &mut expr.kind {
            ast::ExpressionKind::Boolean(_) => ast::Type::Boolean,
            ast::ExpressionKind::Integer(_) => ast::Type::I64,
            ast::ExpressionKind::String(_) => ast::Type::String,
            ast::ExpressionKind::Symbol(ref mut e) => self.analyze_symbol_expression(e),
            ast::ExpressionKind::Binary(ref mut e) => self.analyze_binary_expression(e),
            ast::ExpressionKind::Comparison(ref mut e) => self.analyze_comparison_expression(e),
            ast::ExpressionKind::Unary(ref mut e) => self.analyze_unary_expression(e),
            ast::ExpressionKind::Call(ref mut e) => self.analyze_call_expression(e),
            ast::ExpressionKind::Index(ref mut e) => self.analyze_index_expression(e),
            ast::ExpressionKind::TupleIndex(ref mut e) => self.analyze_tuple_index_expression(e),
            ast::ExpressionKind::Attribute(ref mut e) => self.analyze_attribute_expression(e),
            ast::ExpressionKind::List(ref mut e) => self.analyze_list_literal(e),
            ast::ExpressionKind::Tuple(ref mut e) => self.analyze_tuple_literal(e),
            ast::ExpressionKind::Map(ref mut e) => self.analyze_map_literal(e),
            ast::ExpressionKind::Record(ref mut e) => self.analyze_record_literal(e),
        };
        expr.semantic_type.clone()
    }

    fn analyze_symbol_expression(&mut self, expr: &mut ast::SymbolExpression) -> ast::Type {
        if let Some(entry) = self.symbols.get(&expr.name) {
            expr.entry = Some(entry.clone());
            entry.type_
        } else {
            self.error("unknown symbol", expr.location.clone());
            ast::Type::Error
        }
    }

    fn analyze_binary_expression(&mut self, expr: &mut ast::BinaryExpression) -> ast::Type {
        let left_type = self.analyze_expression(&mut expr.left);
        let right_type = self.analyze_expression(&mut expr.right);
        match expr.op {
            ast::BinaryOpType::Concat => match left_type {
                ast::Type::String => {
                    if !right_type.matches(&ast::Type::String) {
                        self.error_type_mismatch(
                            &ast::Type::String,
                            &right_type,
                            expr.right.location.clone(),
                        );
                        ast::Type::Error
                    } else {
                        ast::Type::String
                    }
                }
                ast::Type::List(ref t) => {
                    if !left_type.matches(&right_type) {
                        self.error_type_mismatch(
                            &left_type,
                            &right_type,
                            expr.right.location.clone(),
                        );
                        ast::Type::Error
                    } else {
                        left_type.clone()
                    }
                }
                _ => {
                    let msg = format!("cannot concatenate value of type {}", left_type);
                    self.error(&msg, expr.left.location.clone());
                    ast::Type::Error
                }
            },
            ast::BinaryOpType::Or | ast::BinaryOpType::And => {
                self.assert_type(&left_type, &ast::Type::Boolean, expr.left.location.clone());
                self.assert_type(
                    &right_type,
                    &ast::Type::Boolean,
                    expr.right.location.clone(),
                );
                ast::Type::Boolean
            }
            _ => {
                self.assert_type(&left_type, &ast::Type::I64, expr.left.location.clone());
                self.assert_type(&right_type, &ast::Type::I64, expr.right.location.clone());
                ast::Type::I64
            }
        }
    }

    fn analyze_comparison_expression(&mut self, expr: &mut ast::ComparisonExpression) -> ast::Type {
        let left_type = self.analyze_expression(&mut expr.left);
        let right_type = self.analyze_expression(&mut expr.right);
        match expr.op {
            ast::ComparisonOpType::Equals | ast::ComparisonOpType::NotEquals => {
                self.assert_type(&left_type, &right_type, expr.left.location.clone());
                ast::Type::Boolean
            }
            ast::ComparisonOpType::LessThan
            | ast::ComparisonOpType::LessThanEquals
            | ast::ComparisonOpType::GreaterThan
            | ast::ComparisonOpType::GreaterThanEquals => {
                self.assert_type(&left_type, &ast::Type::I64, expr.left.location.clone());
                self.assert_type(&right_type, &ast::Type::I64, expr.right.location.clone());
                ast::Type::Boolean
            }
        }
    }

    fn analyze_unary_expression(&mut self, expr: &mut ast::UnaryExpression) -> ast::Type {
        let operand_type = self.analyze_expression(&mut expr.operand);
        match expr.op {
            ast::UnaryOpType::Negate => {
                self.assert_type(
                    &operand_type,
                    &ast::Type::I64,
                    expr.operand.location.clone(),
                );
                ast::Type::I64
            }
            ast::UnaryOpType::Not => {
                self.assert_type(
                    &operand_type,
                    &ast::Type::Boolean,
                    expr.operand.location.clone(),
                );
                ast::Type::Boolean
            }
        }
    }

    fn analyze_call_expression(&mut self, expr: &mut ast::CallExpression) -> ast::Type {
        if let Some(entry) = self.symbols.get(&expr.function.name) {
            if let ast::Type::Function {
                parameters,
                return_type,
            } = &entry.type_
            {
                expr.function.entry = Some(entry.clone());
                if parameters.len() != expr.arguments.len() {
                    let msg = format!(
                        "expected {} parameter(s), got {}",
                        parameters.len(),
                        expr.arguments.len()
                    );
                    self.error(&msg, expr.location.clone());
                }

                for argument in &mut expr.arguments {
                    self.analyze_expression(argument);
                }

                for (parameter, argument) in parameters.iter().zip(expr.arguments.iter()) {
                    self.assert_type(
                        &parameter,
                        &argument.semantic_type,
                        argument.location.clone(),
                    );
                }

                *return_type.clone()
            } else {
                let msg = format!("cannot call non-function type {}", entry.type_);
                self.error(&msg, expr.location.clone());
                ast::Type::Error
            }
        } else {
            let msg = format!("unknown symbol {}", expr.function);
            self.error(&msg, expr.location.clone());
            ast::Type::Error
        }
    }

    fn analyze_index_expression(&mut self, expr: &mut ast::IndexExpression) -> ast::Type {
        let value_type = self.analyze_expression(&mut expr.value);
        let index_type = self.analyze_expression(&mut expr.index);

        match value_type {
            ast::Type::List(t) => {
                self.assert_type(&index_type, &ast::Type::I64, expr.index.location.clone());
                *t.clone()
            }
            ast::Type::Map { key, value } => {
                self.assert_type(&index_type, &key, expr.index.location.clone());
                *value.clone()
            }
            _ => {
                let msg = format!("cannot index non-list, non-map type {}", value_type);
                self.error(&msg, expr.value.location.clone());
                ast::Type::Error
            }
        }
    }

    fn analyze_tuple_index_expression(
        &mut self,
        expr: &mut ast::TupleIndexExpression,
    ) -> ast::Type {
        let value_type = self.analyze_expression(&mut expr.value);
        if let ast::Type::Tuple(ts) = value_type {
            if expr.index >= ts.len() {
                self.error("tuple index out of range", expr.location.clone());
                ast::Type::Error
            } else {
                ts[expr.index].clone()
            }
        } else {
            let msg = format!("cannot index non-tuple type {}", value_type);
            self.error(&msg, expr.location.clone());
            ast::Type::Error
        }
    }

    fn analyze_attribute_expression(&mut self, expr: &mut ast::AttributeExpression) -> ast::Type {
        // TODO
        ast::Type::Error
    }

    fn analyze_list_literal(&mut self, expr: &mut ast::ListLiteral) -> ast::Type {
        if expr.items.len() == 0 {
            self.error(
                "cannot type-check empty list literal",
                expr.location.clone(),
            );
            return ast::Type::Error;
        }

        let item_type = self.analyze_expression(&mut expr.items[0]);
        for i in 1..expr.items.len() {
            let another_item_type = self.analyze_expression(&mut expr.items[i]);
            self.assert_type(
                &another_item_type,
                &item_type,
                expr.items[i].location.clone(),
            );
        }
        ast::Type::List(Box::new(item_type))
    }

    fn analyze_tuple_literal(&mut self, expr: &mut ast::TupleLiteral) -> ast::Type {
        let mut types = Vec::new();
        for item in &mut expr.items {
            types.push(self.analyze_expression(item));
        }
        ast::Type::Tuple(types)
    }

    fn analyze_map_literal(&mut self, expr: &mut ast::MapLiteral) -> ast::Type {
        if expr.items.len() == 0 {
            self.error("cannot type-check empty map literal", expr.location.clone());
            return ast::Type::Error;
        }

        let key_type = self.analyze_expression(&mut expr.items[0].0);
        let value_type = self.analyze_expression(&mut expr.items[0].1);
        for i in 1..expr.items.len() {
            let another_key_type = self.analyze_expression(&mut expr.items[i].0);
            self.assert_type(
                &another_key_type,
                &key_type,
                expr.items[i].0.location.clone(),
            );
            let another_value_type = self.analyze_expression(&mut expr.items[i].1);
            self.assert_type(
                &another_value_type,
                &value_type,
                expr.items[i].1.location.clone(),
            );
        }
        ast::Type::Map {
            key: Box::new(key_type),
            value: Box::new(value_type),
        }
    }

    fn analyze_record_literal(&mut self, expr: &mut ast::RecordLiteral) -> ast::Type {
        // TODO
        ast::Type::Error
    }

    fn resolve_type(&mut self, type_: &ast::SyntacticType) -> ast::Type {
        match &type_.kind {
            ast::SyntacticTypeKind::Literal(s) => {
                if let Some(semantic_type_) = self.types.get(&s) {
                    semantic_type_.type_
                } else {
                    let msg = format!("unknown type {}", s);
                    self.error(&msg, type_.location.clone());
                    ast::Type::Error
                }
            }
            _ => {
                // TODO
                ast::Type::Error
            }
        }
    }

    fn claim_unique_name(&mut self, prefix: &str) -> String {
        let c = self.unique_name_counter;
        self.unique_name_counter += 1;
        format!("{}__{}", prefix, c)
    }

    fn assert_type(
        &mut self,
        actual: &ast::Type,
        expected: &ast::Type,
        location: common::Location,
    ) {
        if !actual.matches(expected) {
            self.error_type_mismatch(expected, actual, location);
        }
    }

    fn error(&mut self, message: &str, location: common::Location) {
        self.errors
            .push(errors::VeniceError::new(message, location));
    }

    fn error_type_mismatch(
        &mut self,
        expected: &ast::Type,
        actual: &ast::Type,
        location: common::Location,
    ) {
        let msg = format!("expected {}, got {}", expected, actual);
        self.error(&msg, location);
    }
}
