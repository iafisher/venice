// Copyright 2022 The Venice Authors. All rights reserved.
// Use of this source code is governed by an MIT-style license that can be
// found in the LICENSE file.

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
