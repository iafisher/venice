#[derive(Clone)]
pub struct VeniceError {
    pub message: String,
}

impl VeniceError {
    pub fn new(message: &str) -> Self {
        VeniceError {
            message: String::from(message),
        }
    }
}
