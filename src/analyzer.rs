// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.
//
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

struct Analyzer {
    symbols: SymbolTable,
    types: SymbolTable,
    current_function_return_type: Option<ast::Type>,
    current_function_info: Option<ast::FunctionInfo>,
    errors: Vec<errors::VeniceError>,
    unique_name_counter: u64,
    current_stack_offset: i32,
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
            current_stack_offset: 0,
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
        let mut stack_frame_size = 0;
        let mut stack_offset = -8;
        for parameter in &declaration.parameters {
            let t = self.resolve_type(&parameter.type_);
            let unique_name = self.claim_unique_name(&parameter.name);
            let entry = ast::SymbolEntry {
                unique_name,
                type_: t.clone(),
                constant: false,
                external: false,
                stack_offset,
            };

            stack_frame_size += t.stack_size();
            stack_offset -= t.stack_size() as i32;
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

        // Add the function's entry to the symbol table. It must happen before we push a new scope
        // so that the function's definition persists after it executes. It also must happen before
        // we type-check the body of the function so that recursive functions work.
        let entry = ast::SymbolEntry {
            unique_name,
            type_: ast::Type::Function {
                parameters: parameter_types,
                return_type: Box::new(return_type.clone()),
            },
            constant: true,
            external: false,
            stack_offset: 0,
        };
        self.symbols.insert(&declaration.name, entry.clone());

        // Push a new scope for the function's body and add the parameters.
        self.symbols.push_scope();
        for (ptree_parameter, ast_parameter) in declaration.parameters.iter().zip(&parameters) {
            self.symbols
                .insert(&ptree_parameter.name, ast_parameter.name.clone());
        }

        self.current_function_info = Some(ast::FunctionInfo { stack_frame_size });
        self.current_function_return_type = Some(return_type.clone());
        self.current_stack_offset = stack_offset;
        let body = self.analyze_block(&declaration.body);
        self.current_function_return_type = None;
        self.current_stack_offset = -8;

        // Pop off the function body's scope.
        self.symbols.pop_scope();

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
            stack_offset: 0,
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
            stack_offset: self.current_stack_offset,
        };

        self.symbols.insert(&stmt.symbol, entry.clone());
        self.current_function_info
            .as_mut()
            .unwrap()
            .stack_frame_size += entry.type_.stack_size();
        self.current_stack_offset -= declared_type.stack_size() as i32;

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
                        variadic: false,
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
                    kind: ast::ExpressionKind::Call(ast::CallExpression {
                        function: ast::SymbolEntry {
                            unique_name: String::from("venice_list_index"),
                            type_: ast::Type::Error,
                            constant: true,
                            external: true,
                            stack_offset: 0,
                        },
                        arguments: vec![value, index],
                        variadic: false,
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

        let mut arguments = Vec::new();
        arguments.push(ast::Expression {
            kind: ast::ExpressionKind::Integer(expr.items.len() as i64),
            type_: ast::Type::I64,
        });

        let first_item = self.analyze_expression(&expr.items[0]);
        let item_type = first_item.type_.clone();
        arguments.push(first_item);

        for i in 1..expr.items.len() {
            let typed_item = self.analyze_expression(&expr.items[i]);
            self.assert_type(
                &typed_item.type_,
                &item_type,
                expr.items[i].location.clone(),
            );
            arguments.push(typed_item);
        }
        ast::Expression {
            kind: ast::ExpressionKind::Call(ast::CallExpression {
                function: ast::SymbolEntry {
                    unique_name: String::from("venice_list_new"),
                    type_: ast::Type::Error,
                    constant: true,
                    external: true,
                    stack_offset: 0,
                },
                arguments,
                variadic: true,
            }),
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

    fn resolve_type(&mut self, type_: &ptree::Type) -> ast::Type {
        match &type_.kind {
            ptree::TypeKind::Literal(s) => {
                if let Some(entry) = self.types.get(s) {
                    entry.type_
                } else {
                    let msg = format!("unknown type {}", s);
                    self.error(&msg, type_.location.clone());
                    ast::Type::Error
                }
            }
            ptree::TypeKind::Parameterized(ptree::ParameterizedType { symbol, parameters }) => {
                if symbol == "list" {
                    if parameters.len() == 1 {
                        let item_type = self.resolve_type(&parameters[0]);
                        ast::Type::List(Box::new(item_type))
                    } else {
                        self.error(
                            "expected 1 type parameter to 'list'",
                            type_.location.clone(),
                        );
                        ast::Type::Error
                    }
                } else {
                    let msg = format!("unknown type {}", symbol);
                    self.error(&msg, type_.location.clone());
                    ast::Type::Error
                }
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

struct SymbolTable {
    environments: Vec<HashMap<String, ast::SymbolEntry>>,
}

impl SymbolTable {
    pub fn new() -> Self {
        SymbolTable {
            environments: vec![HashMap::new()],
        }
    }

    pub fn builtin_types() -> Self {
        let mut symbols = HashMap::new();
        symbols.insert(String::from("i64"), ast::SymbolEntry::type_(ast::Type::I64));
        symbols.insert(
            String::from("bool"),
            ast::SymbolEntry::type_(ast::Type::Boolean),
        );
        symbols.insert(
            String::from("string"),
            ast::SymbolEntry::type_(ast::Type::String),
        );
        symbols.insert(
            String::from("void"),
            ast::SymbolEntry::type_(ast::Type::Void),
        );

        SymbolTable {
            environments: vec![symbols],
        }
    }

    pub fn builtin_globals() -> Self {
        let mut symbols = HashMap::new();
        // TODO: unique names here could conflict with actual user symbols.
        symbols.insert(
            String::from("println"),
            ast::SymbolEntry::external(
                "venice_println",
                ast::Type::Function {
                    parameters: vec![ast::Type::String],
                    return_type: Box::new(ast::Type::Void),
                },
            ),
        );
        // TODO: remove printint once there's a better way to print integers.
        symbols.insert(
            String::from("printint"),
            ast::SymbolEntry::external(
                "venice_printint",
                ast::Type::Function {
                    parameters: vec![ast::Type::I64],
                    return_type: Box::new(ast::Type::Void),
                },
            ),
        );

        SymbolTable {
            environments: vec![symbols],
        }
    }

    pub fn get(&self, key: &str) -> Option<ast::SymbolEntry> {
        for table in self.environments.iter().rev() {
            if let Some(entry) = table.get(key) {
                return Some(entry.clone());
            }
        }
        None
    }

    pub fn insert(&mut self, key: &str, entry: ast::SymbolEntry) {
        self.current().insert(String::from(key), entry);
    }

    pub fn remove(&mut self, key: &str) {
        self.current().remove(key);
    }

    pub fn push_scope(&mut self) {
        self.environments.push(HashMap::new());
    }

    pub fn pop_scope(&mut self) {
        self.environments.pop();
    }

    fn current(&mut self) -> &mut HashMap<String, ast::SymbolEntry> {
        let index = self.environments.len() - 1;
        &mut self.environments[index]
    }
}
