use super::ast;
use super::errors;
use super::lexer;

pub fn parse(lexer: lexer::Lexer) -> Result<ast::Program, Vec<errors::VeniceError>> {
    let mut parser = Parser::new(lexer);
    let ast = parser.parse();
    if parser.errors.len() > 0 {
        Err(parser.errors.clone())
    } else {
        Ok(ast)
    }
}

struct Parser {
    lexer: lexer::Lexer,
    errors: Vec<errors::VeniceError>,
}

impl Parser {
    fn new(lexer: lexer::Lexer) -> Self {
        Parser {
            lexer: lexer,
            errors: Vec::new(),
        }
    }

    fn parse(&mut self) -> ast::Program {
        ast::Program {
            declarations: Vec::new(),
        }
    }
}
