mod analyzer;
mod ast;
mod codegen;
mod common;
mod lexer;
mod ptree;
mod vil;
mod x86;

fn main() {
    let n = ptree::Node {
        name: String::from("whatever"),
    };
    println!("{:?}", n);
}
