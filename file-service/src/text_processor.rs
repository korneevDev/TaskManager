use std::{borrow::Cow, fs, path::Path};
use thiserror::Error;

#[derive(Error, Debug)]
pub enum TextProcessingError {
    #[error("File not found")]
    FileNotFound,
    #[error("File read error")]
    FileReadError,
}

pub struct TextProcessor<'a> {
    content: Cow<'a, str>,
}

impl<'a> TextProcessor<'a> {
    pub fn new(file_path: &'a Path) -> Result<Self, TextProcessingError> {
        if !file_path.exists() {
            return Err(TextProcessingError::FileNotFound);
        }

        let content = fs::read_to_string(file_path)
            .map_err(|_| TextProcessingError::FileReadError)?;
        
        Ok(Self {
            content: Cow::Owned(content),
        })
    }

    pub fn content(&self) -> &str {
        &self.content
    }

    pub fn count_words(&self) -> usize {
        self.content.split_whitespace().count()
    }

    pub fn count_word_occurrences(&self, word: &str) -> usize {
        self.content
            .split_whitespace()
            .filter(|&w| w == word)
            .count()
    }

    pub fn find_lines_with<'b>(&'b self, substring: &str) -> Vec<&'b str> {
        self.content
            .lines()
            .filter(|line| line.contains(substring))
            .collect()
    }
}

pub fn process_text_owned(text: String) -> String {
    text.to_uppercase()
}

pub fn process_text_borrowed(text: &str) -> String {
    text.to_lowercase()
}