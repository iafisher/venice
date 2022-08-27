// The analyzer turns the parse tree into an abstract syntax tree by computing type information and
// simplifying syntactic sugar.

use super::ast;
use super::common;
use super::errors;
use super::ptree;
use std::collections::HashMap;

/// Analyzes the parse tree into an abstract syntax tree.
pub fn analyze(ptree: &ptree::Program) -> Result<ast::Program, Vec<errors::VeniceError>> {
    let mut analyzer = Analyzer::new();
    let program = analyzer.analyze_program(ptree);
    if !analyzer.errors.is_empty() {
        Err(analyzer.errors.clone())
    } else {
        Ok(program)
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

        SymbolTable { symbols }
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
        self.symbols.insert(String::from(key), entry);
    }

    pub fn remove(&mut self, key: &str) {
        self.symbols.remove(key);
    }
}

struct Analyzer {
    symbols: SymbolTable,
    types: SymbolTable,
    current_function_return_type: Option<ast::Type>,
    current_function_info: Option<ast::FunctionInfo>,
    errors: Vec<errors::VeniceError>,
    unique_name_counter: u64,
}

impl Analyzer {
    fn new() -> Self {
        Analyzer {
            symbols: SymbolTable::builtin_globals(),
            types: SymbolTable::builtin_types(),
            current_function_return_type: None,
            current_function_info: None,
            errors: Vec::new(),
            unique_name_counter: 0,
        }
    }

    fn analyze_program(&mut self, ptree: &ptree::Program) -> ast::Program {
        let mut declarations = Vec::new();
        for declaration in &ptree.declarations {
            declarations.push(self.analyze_declaration(declaration));
        }
        ast::Program { declarations }
    }

    fn analyze_declaration(&mut self, declaration: &ptree::Declaration) -> ast::Declaration {
        match declaration {
            ptree::Declaration::Function(d) => self.analyze_function_declaration(d),
            ptree::Declaration::Const(d) => self.analyze_const_declaration(d),
            ptree::Declaration::Record(d) => self.analyze_record_declaration(d),
        }
    }

    fn analyze_function_declaration(
        &mut self,
        declaration: &ptree::FunctionDeclaration,
    ) -> ast::Declaration {
        let return_type = self.resolve_type(&declaration.return_type);
        let mut parameters = Vec::new();
        let mut parameter_types = Vec::new();
        for parameter in &declaration.parameters {
            let t = self.resolve_type(&parameter.type_);
            let unique_name = self.claim_unique_name(&parameter.name);
            let entry = ast::SymbolEntry {
                unique_name,
                type_: t.clone(),
                constant: false,
                external: false,
            };
            self.symbols.insert(&parameter.name, entry.clone());
            parameter_types.push(t.clone());
            parameters.push(ast::FunctionParameter {
                name: entry,
                type_: t,
            });
        }

        let unique_name = if declaration.name == "main" {
            // Keep main's name the same so that the linker can find it.
            String::from("main")
        } else {
            self.claim_unique_name(&declaration.name)
        };

        let entry = ast::SymbolEntry {
            unique_name,
            type_: ast::Type::Function {
                parameters: parameter_types,
                return_type: Box::new(return_type.clone()),
            },
            constant: true,
            external: false,
        };

        self.symbols.insert(&declaration.name, entry.clone());

        self.current_function_info = Some(ast::FunctionInfo {
            // TODO: not all parameters are 8 bytes
            stack_frame_size: 8 * declaration.parameters.len(),
        });
        self.current_function_return_type = Some(return_type.clone());
        let body = self.analyze_block(&declaration.body);
        self.current_function_return_type = None;

        // TODO: A better approach would be to have a separate symbol table for the
        // function's scope. And actually this doesn't work at all because the symbols
        // defined in a function's body will persist after the function is finished.
        for parameter in &declaration.parameters {
            self.symbols.remove(&parameter.name);
        }

        ast::Declaration::Function(ast::FunctionDeclaration {
            name: entry,
            parameters,
            return_type,
            body,
            info: self.current_function_info.as_ref().unwrap().clone(),
        })
    }

    fn analyze_const_declaration(
        &mut self,
        declaration: &ptree::ConstDeclaration,
    ) -> ast::Declaration {
        let value = self.analyze_expression(&declaration.value);
        let declared_type = self.resolve_type(&declaration.type_);
        if !declared_type.matches(&value.type_) {
            self.error_type_mismatch(&declared_type, &value.type_, declaration.location.clone());
        }

        let unique_name = self.claim_unique_name(&declaration.symbol);
        let entry = ast::SymbolEntry {
            unique_name,
            type_: declared_type.clone(),
            constant: true,
            external: false,
        };
        self.symbols.insert(&declaration.symbol, entry.clone());

        ast::Declaration::Const(ast::ConstDeclaration {
            symbol: entry,
            type_: declared_type,
            value,
        })
    }

    fn analyze_record_declaration(
        &mut self,
        declaration: &ptree::RecordDeclaration,
    ) -> ast::Declaration {
        // TODO
        self.error("not implemented", declaration.location.clone());
        ast::Declaration::Error
    }

    fn analyze_block(&mut self, block: &[ptree::Statement]) -> Vec<ast::Statement> {
        let mut ret = Vec::new();
        for stmt in block {
            ret.push(self.analyze_statement(stmt));
        }
        ret
    }

    fn analyze_statement(&mut self, stmt: &ptree::Statement) -> ast::Statement {
        match stmt {
            ptree::Statement::Let(s) => self.analyze_let_statement(s),
            ptree::Statement::Assign(s) => self.analyze_assign_statement(s),
            ptree::Statement::If(s) => self.analyze_if_statement(s),
            ptree::Statement::While(s) => self.analyze_while_statement(s),
            ptree::Statement::For(s) => self.analyze_for_statement(s),
            ptree::Statement::Return(s) => self.analyze_return_statement(s),
            ptree::Statement::Assert(s) => self.analyze_assert_statement(s),
            ptree::Statement::Expression(expr) => {
                ast::Statement::Expression(self.analyze_expression(expr))
            }
        }
    }

    fn analyze_let_statement(&mut self, stmt: &ptree::LetStatement) -> ast::Statement {
        let value = self.analyze_expression(&stmt.value);
        let declared_type = self.resolve_type(&stmt.type_);
        if !declared_type.matches(&value.type_) {
            self.error_type_mismatch(&declared_type, &value.type_, stmt.location.clone());
        }

        let unique_name = self.claim_unique_name(&stmt.symbol);
        let entry = ast::SymbolEntry {
            unique_name,
            type_: declared_type.clone(),
            constant: false,
            external: false,
        };

        self.symbols.insert(&stmt.symbol, entry.clone());
        // TODO: not all symbols are 8 bytes
        self.current_function_info
            .as_mut()
            .unwrap()
            .stack_frame_size += 8;

        ast::Statement::Let(ast::LetStatement {
            symbol: entry,
            type_: declared_type,
            value,
        })
    }

    fn analyze_assign_statement(&mut self, stmt: &ptree::AssignStatement) -> ast::Statement {
        let value = self.analyze_expression(&stmt.value);
        if let Some(entry) = self.symbols.get(&stmt.symbol) {
            if !entry.type_.matches(&value.type_) {
                self.error_type_mismatch(&entry.type_, &value.type_, stmt.location.clone());
            }
            ast::Statement::Assign(ast::AssignStatement {
                symbol: entry.clone(),
                value,
            })
        } else {
            let msg = format!("assignment to unknown symbol {}", stmt.symbol);
            self.error(&msg, stmt.location.clone());
            ast::Statement::Error
        }
    }

    fn analyze_if_statement(&mut self, stmt: &ptree::IfStatement) -> ast::Statement {
        let condition = self.analyze_expression(&stmt.if_clause.condition);
        if !condition.type_.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(
                &ast::Type::Boolean,
                &condition.type_,
                stmt.if_clause.condition.location.clone(),
            );
        }
        let body = self.analyze_block(&stmt.if_clause.body);
        let else_body = self.analyze_block(&stmt.else_body);

        if !stmt.elif_clauses.is_empty() {
            self.error(
                "not implemented",
                stmt.elif_clauses[0].condition.location.clone(),
            );
            ast::Statement::Error
            /*
            for elif_clause in &stmt.elif_clauses {
                let elif_condition = self.analyze_expression(&elif_clause.condition)?;
                if !elif_condition.type_.matches(&ast::Type::Boolean) {
                    self.error_type_mismatch(
                        &ast::Type::Boolean,
                        &elif_condition.type_,
                        elif_clause.condition.location.clone(),
                    );
                }
                let elif_body = self.analyze_block(&mut elif_clause.body)?;
            }
            */
        } else {
            ast::Statement::If(ast::IfStatement {
                condition,
                body,
                else_body,
            })
        }
    }

    fn analyze_while_statement(&mut self, stmt: &ptree::WhileStatement) -> ast::Statement {
        let condition = self.analyze_expression(&stmt.condition);
        if !condition.type_.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(
                &ast::Type::Boolean,
                &condition.type_,
                stmt.condition.location.clone(),
            );
        }
        let body = self.analyze_block(&stmt.body);
        ast::Statement::While(ast::WhileStatement { condition, body })
    }

    fn analyze_for_statement(&mut self, stmt: &ptree::ForStatement) -> ast::Statement {
        // TODO
        self.error("not implemented", stmt.location.clone());
        ast::Statement::Error
    }

    fn analyze_return_statement(&mut self, stmt: &ptree::ReturnStatement) -> ast::Statement {
        let value = self.analyze_expression(&stmt.value);
        // TODO: Can the clone here be avoided?
        if let Some(expected_return_type) = self.current_function_return_type.clone() {
            if !expected_return_type.matches(&value.type_) {
                self.error_type_mismatch(
                    &expected_return_type,
                    &value.type_,
                    stmt.location.clone(),
                );
            }
        } else {
            self.error(
                "return statement outside of function",
                stmt.location.clone(),
            );
        }
        ast::Statement::Return(ast::ReturnStatement { value })
    }

    fn analyze_assert_statement(&mut self, stmt: &ptree::AssertStatement) -> ast::Statement {
        let condition = self.analyze_expression(&stmt.condition);
        if !condition.type_.matches(&ast::Type::Boolean) {
            self.error_type_mismatch(
                &ast::Type::Boolean,
                &condition.type_,
                stmt.condition.location.clone(),
            );
        }
        ast::Statement::Assert(ast::AssertStatement { condition })
    }

    fn analyze_expression(&mut self, expr: &ptree::Expression) -> ast::Expression {
        match &expr.kind {
            ptree::ExpressionKind::Boolean(x) => ast::Expression {
                kind: ast::ExpressionKind::Boolean(*x),
                type_: ast::Type::Boolean,
            },
            ptree::ExpressionKind::Integer(x) => ast::Expression {
                kind: ast::ExpressionKind::Integer(*x),
                type_: ast::Type::I64,
            },
            ptree::ExpressionKind::String(x) => ast::Expression {
                kind: ast::ExpressionKind::String(x.clone()),
                type_: ast::Type::String,
            },
            ptree::ExpressionKind::Symbol(ref e) => self.analyze_symbol(e, &expr.location),
            ptree::ExpressionKind::Binary(ref e) => self.analyze_binary_expression(e),
            ptree::ExpressionKind::Comparison(ref e) => self.analyze_comparison_expression(e),
            ptree::ExpressionKind::Unary(ref e) => self.analyze_unary_expression(e),
            ptree::ExpressionKind::Call(ref e) => self.analyze_call_expression(e),
            ptree::ExpressionKind::Index(ref e) => self.analyze_index_expression(e),
            ptree::ExpressionKind::TupleIndex(ref e) => self.analyze_tuple_index_expression(e),
            ptree::ExpressionKind::Attribute(ref e) => self.analyze_attribute_expression(e),
            ptree::ExpressionKind::List(ref e) => self.analyze_list_literal(e),
            ptree::ExpressionKind::Tuple(ref e) => self.analyze_tuple_literal(e),
            ptree::ExpressionKind::Map(ref e) => self.analyze_map_literal(e),
            ptree::ExpressionKind::Record(ref e) => self.analyze_record_literal(e),
        }
    }

    fn analyze_symbol(&mut self, name: &str, location: &common::Location) -> ast::Expression {
        if let Some(entry) = self.symbols.get(name) {
            ast::Expression {
                kind: ast::ExpressionKind::Symbol(entry.clone()),
                type_: entry.type_,
            }
        } else {
            self.error("unknown symbol", location.clone());
            ast::EXPRESSION_ERROR.clone()
        }
    }

    fn analyze_binary_expression(&mut self, expr: &ptree::BinaryExpression) -> ast::Expression {
        let left = self.analyze_expression(&expr.left);
        let right = self.analyze_expression(&expr.right);
        match expr.op {
            common::BinaryOpType::Concat => match &left.type_ {
                ast::Type::String => {
                    if !right.type_.matches(&ast::Type::String) {
                        self.error_type_mismatch(
                            &ast::Type::String,
                            &right.type_,
                            expr.right.location.clone(),
                        );
                        ast::EXPRESSION_ERROR.clone()
                    } else {
                        ast::Expression {
                            kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                op: common::BinaryOpType::Concat,
                                left: Box::new(left),
                                right: Box::new(right),
                            }),
                            type_: ast::Type::String,
                        }
                    }
                }
                ast::Type::List(_) => {
                    if !left.type_.matches(&right.type_) {
                        self.error_type_mismatch(
                            &left.type_,
                            &right.type_,
                            expr.right.location.clone(),
                        );
                        ast::EXPRESSION_ERROR.clone()
                    } else {
                        let type_ = left.type_.clone();
                        ast::Expression {
                            kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                                op: common::BinaryOpType::Concat,
                                left: Box::new(left),
                                right: Box::new(right),
                            }),
                            type_,
                        }
                    }
                }
                _ => {
                    let msg = format!("cannot concatenate value of type {}", left.type_);
                    self.error(&msg, expr.left.location.clone());
                    ast::EXPRESSION_ERROR.clone()
                }
            },
            common::BinaryOpType::Or | common::BinaryOpType::And => {
                self.assert_type(&left.type_, &ast::Type::Boolean, expr.left.location.clone());
                self.assert_type(
                    &right.type_,
                    &ast::Type::Boolean,
                    expr.right.location.clone(),
                );
                ast::Expression {
                    kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                        op: expr.op,
                        left: Box::new(left),
                        right: Box::new(right),
                    }),
                    type_: ast::Type::Boolean,
                }
            }
            _ => {
                self.assert_type(&left.type_, &ast::Type::I64, expr.left.location.clone());
                self.assert_type(&right.type_, &ast::Type::I64, expr.right.location.clone());
                ast::Expression {
                    kind: ast::ExpressionKind::Binary(ast::BinaryExpression {
                        op: expr.op,
                        left: Box::new(left),
                        right: Box::new(right),
                    }),
                    type_: ast::Type::I64,
                }
            }
        }
    }

    fn analyze_comparison_expression(
        &mut self,
        expr: &ptree::ComparisonExpression,
    ) -> ast::Expression {
        let left = self.analyze_expression(&expr.left);
        let right = self.analyze_expression(&expr.right);
        match expr.op {
            common::ComparisonOpType::Equals | common::ComparisonOpType::NotEquals => {
                self.assert_type(&left.type_, &right.type_, expr.left.location.clone());
                ast::Expression {
                    kind: ast::ExpressionKind::Comparison(ast::ComparisonExpression {
                        op: expr.op,
                        left: Box::new(left),
                        right: Box::new(right),
                    }),
                    type_: ast::Type::Boolean,
                }
            }
            common::ComparisonOpType::LessThan
            | common::ComparisonOpType::LessThanEquals
            | common::ComparisonOpType::GreaterThan
            | common::ComparisonOpType::GreaterThanEquals => {
                self.assert_type(&left.type_, &ast::Type::I64, expr.left.location.clone());
                self.assert_type(&right.type_, &ast::Type::I64, expr.right.location.clone());
                ast::Expression {
                    kind: ast::ExpressionKind::Comparison(ast::ComparisonExpression {
                        op: expr.op,
                        left: Box::new(left),
                        right: Box::new(right),
                    }),
                    type_: ast::Type::Boolean,
                }
            }
        }
    }

    fn analyze_unary_expression(&mut self, expr: &ptree::UnaryExpression) -> ast::Expression {
        let operand = self.analyze_expression(&expr.operand);
        match expr.op {
            common::UnaryOpType::Negate => {
                self.assert_type(
                    &operand.type_,
                    &ast::Type::I64,
                    expr.operand.location.clone(),
                );
                ast::Expression {
                    kind: ast::ExpressionKind::Unary(ast::UnaryExpression {
                        op: expr.op,
                        operand: Box::new(operand),
                    }),
                    type_: ast::Type::I64,
                }
            }
            common::UnaryOpType::Not => {
                self.assert_type(
                    &operand.type_,
                    &ast::Type::Boolean,
                    expr.operand.location.clone(),
                );
                ast::Expression {
                    kind: ast::ExpressionKind::Unary(ast::UnaryExpression {
                        op: expr.op,
                        operand: Box::new(operand),
                    }),
                    type_: ast::Type::I64,
                }
            }
        }
    }

    fn analyze_call_expression(&mut self, expr: &ptree::CallExpression) -> ast::Expression {
        if let Some(entry) = self.symbols.get(&expr.function) {
            if let ast::Type::Function {
                parameters,
                return_type,
            } = &entry.type_
            {
                if parameters.len() != expr.arguments.len() {
                    let msg = format!(
                        "expected {} parameter(s), got {}",
                        parameters.len(),
                        expr.arguments.len()
                    );
                    self.error(&msg, expr.location.clone());
                }

                let mut arguments = Vec::new();
                for (parameter, argument) in parameters.iter().zip(expr.arguments.iter()) {
                    let typed_argument = self.analyze_expression(argument);
                    self.assert_type(parameter, &typed_argument.type_, argument.location.clone());
                    arguments.push(typed_argument);
                }

                ast::Expression {
                    kind: ast::ExpressionKind::Call(ast::CallExpression {
                        function: entry.clone(),
                        arguments,
                    }),
                    type_: *return_type.clone(),
                }
            } else {
                let msg = format!("cannot call non-function type {}", entry.type_);
                self.error(&msg, expr.location.clone());
                ast::EXPRESSION_ERROR.clone()
            }
        } else {
            let msg = format!("unknown symbol {}", expr.function);
            self.error(&msg, expr.location.clone());
            ast::EXPRESSION_ERROR.clone()
        }
    }

    fn analyze_index_expression(&mut self, expr: &ptree::IndexExpression) -> ast::Expression {
        let value = self.analyze_expression(&expr.value);
        let index = self.analyze_expression(&expr.index);

        match &value.type_ {
            ast::Type::List(ref t) => {
                self.assert_type(&index.type_, &ast::Type::I64, expr.index.location.clone());
                let type_ = *t.clone();
                ast::Expression {
                    kind: ast::ExpressionKind::Index(ast::IndexExpression {
                        value: Box::new(value),
                        index: Box::new(index),
                    }),
                    type_,
                }
            }
            ast::Type::Map {
                key: key_type,
                value: ref value_type,
            } => {
                self.assert_type(&index.type_, key_type, expr.index.location.clone());
                let type_ = *value_type.clone();
                ast::Expression {
                    kind: ast::ExpressionKind::Index(ast::IndexExpression {
                        value: Box::new(value),
                        index: Box::new(index),
                    }),
                    type_,
                }
            }
            _ => {
                let msg = format!("cannot index non-list, non-map type {}", value.type_);
                self.error(&msg, expr.value.location.clone());
                ast::EXPRESSION_ERROR.clone()
            }
        }
    }

    fn analyze_tuple_index_expression(
        &mut self,
        expr: &ptree::TupleIndexExpression,
    ) -> ast::Expression {
        let value = self.analyze_expression(&expr.value);
        if let ast::Type::Tuple(ref ts) = &value.type_ {
            if expr.index >= ts.len() {
                self.error("tuple index out of range", expr.location.clone());
                ast::EXPRESSION_ERROR.clone()
            } else {
                let type_ = ts[expr.index].clone();
                ast::Expression {
                    kind: ast::ExpressionKind::TupleIndex(ast::TupleIndexExpression {
                        value: Box::new(value),
                        index: expr.index,
                    }),
                    type_,
                }
            }
        } else {
            let msg = format!("cannot index non-tuple type {}", value.type_);
            self.error(&msg, expr.location.clone());
            ast::EXPRESSION_ERROR.clone()
        }
    }

    fn analyze_attribute_expression(
        &mut self,
        expr: &ptree::AttributeExpression,
    ) -> ast::Expression {
        self.error("not implemented", expr.location.clone());
        ast::EXPRESSION_ERROR.clone()
    }

    fn analyze_list_literal(&mut self, expr: &ptree::ListLiteral) -> ast::Expression {
        if expr.items.is_empty() {
            self.error(
                "cannot type-check empty list literal",
                expr.location.clone(),
            );
            return ast::EXPRESSION_ERROR.clone();
        }

        let first_item = self.analyze_expression(&expr.items[0]);
        let item_type = first_item.type_.clone();
        let mut items = vec![first_item];

        for i in 1..expr.items.len() {
            let typed_item = self.analyze_expression(&expr.items[i]);
            self.assert_type(
                &typed_item.type_,
                &item_type,
                expr.items[i].location.clone(),
            );
            items.push(typed_item);
        }
        ast::Expression {
            kind: ast::ExpressionKind::List(ast::ListLiteral { items }),
            type_: ast::Type::List(Box::new(item_type)),
        }
    }

    fn analyze_tuple_literal(&mut self, expr: &ptree::TupleLiteral) -> ast::Expression {
        let mut items = Vec::new();
        let mut types = Vec::new();
        for item in &expr.items {
            let typed_item = self.analyze_expression(item);
            types.push(typed_item.type_.clone());
            items.push(typed_item);
        }
        ast::Expression {
            kind: ast::ExpressionKind::Tuple(ast::TupleLiteral { items }),
            type_: ast::Type::Tuple(types),
        }
    }

    fn analyze_map_literal(&mut self, expr: &ptree::MapLiteral) -> ast::Expression {
        if expr.items.is_empty() {
            self.error("cannot type-check empty map literal", expr.location.clone());
            return ast::EXPRESSION_ERROR.clone();
        }

        let first_key = self.analyze_expression(&expr.items[0].0);
        let key_type = first_key.type_;
        let first_value = self.analyze_expression(&expr.items[0].1);
        let value_type = first_value.type_;

        let mut items = Vec::new();
        for i in 1..expr.items.len() {
            let typed_key = self.analyze_expression(&expr.items[i].0);
            self.assert_type(
                &typed_key.type_,
                &key_type,
                expr.items[i].0.location.clone(),
            );

            let typed_value = self.analyze_expression(&expr.items[i].1);
            self.assert_type(
                &typed_value.type_,
                &value_type,
                expr.items[i].1.location.clone(),
            );

            items.push((typed_key, typed_value));
        }

        ast::Expression {
            kind: ast::ExpressionKind::Map(ast::MapLiteral { items }),
            type_: ast::Type::Map {
                key: Box::new(key_type),
                value: Box::new(value_type),
            },
        }
    }

    fn analyze_record_literal(&mut self, expr: &ptree::RecordLiteral) -> ast::Expression {
        // TODO
        self.error("not implemented", expr.location.clone());
        ast::EXPRESSION_ERROR.clone()
    }

    fn resolve_type(&mut self, type_: &ptree::SyntacticType) -> ast::Type {
        match &type_.kind {
            ptree::SyntacticTypeKind::Literal(s) => {
                if let Some(semantic_type_) = self.types.get(s) {
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
