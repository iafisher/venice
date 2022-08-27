use super::common;

#[derive(Clone, Debug)]
pub struct VeniceError {
    pub message: String,
    pub location: common::Location,
}

impl VeniceError {
    pub fn new(message: &str, location: common::Location) -> Self {
        VeniceError {
            message: String::from(message),
            location,
        }
    }
}
