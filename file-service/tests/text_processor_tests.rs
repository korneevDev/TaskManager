use file_service::text_processor::{TextProcessor, TextProcessingError, process_text_owned, process_text_borrowed};
use std::io::Write;
use std::path::Path;
use tempfile::NamedTempFile;

// Вспомогательная функция
fn create_temp_file(content: &str) -> NamedTempFile {
    let mut file = NamedTempFile::new().unwrap();
    write!(file, "{}", content).unwrap();
    file
}

#[test]
fn test_new_processor_file_not_found() {
    let path = Path::new("non_existent_file.txt");
    let result = TextProcessor::new(path);
    assert!(matches!(result, Err(TextProcessingError::FileNotFound)));
}

#[test]
fn test_new_processor_success() {
    let file = create_temp_file("Test content");
    let processor = TextProcessor::new(file.path()).unwrap();
    assert_eq!(processor.content(), "Test content");
}

#[test]
fn test_count_words() {
    let file = create_temp_file("Rust is awesome and Rust is fast");
    let processor = TextProcessor::new(file.path()).unwrap();
    assert_eq!(processor.count_words(), 7);
}

#[test]
fn test_count_word_occurrences() {
    let file = create_temp_file("Rust is awesome and Rust is fast");
    let processor = TextProcessor::new(file.path()).unwrap();
    assert_eq!(processor.count_word_occurrences("Rust"), 2);
    assert_eq!(processor.count_word_occurrences("is"), 2);
    assert_eq!(processor.count_word_occurrences("awesome"), 1);
    assert_eq!(processor.count_word_occurrences("missing"), 0);
}

#[test]
fn test_find_lines_with() {
    let content = "First line\nSecond line with Rust\nThird line\nFourth line with Rust and more";
    let file = create_temp_file(content);
    let processor = TextProcessor::new(file.path()).unwrap();
    
    let lines = processor.find_lines_with("Rust");
    assert_eq!(lines.len(), 2);
    assert_eq!(lines[0], "Second line with Rust");
    assert_eq!(lines[1], "Fourth line with Rust and more");
}

#[test]
fn test_process_text_owned() {
    let input = "Hello Rust".to_string();
    let output = process_text_owned(input);
    assert_eq!(output, "HELLO RUST");
}

#[test]
fn test_process_text_borrowed() {
    let input = "Hello Rust";
    let output = process_text_borrowed(input);
    assert_eq!(output, "hello rust");
}

#[test]
fn test_empty_file() {
    let file = create_temp_file("");
    let processor = TextProcessor::new(file.path()).unwrap();
    
    assert_eq!(processor.content(), "");
    assert_eq!(processor.count_words(), 0);
    assert_eq!(processor.count_word_occurrences("any"), 0);
    assert!(processor.find_lines_with("any").is_empty());
}

#[test]
fn test_multiline_content() {
    let content = "Line 1\nLine 2\nLine 3";
    let file = create_temp_file(content);
    let processor = TextProcessor::new(file.path()).unwrap();
    
    assert_eq!(processor.content(), content);
    assert_eq!(processor.count_words(), 6);
    assert_eq!(processor.find_lines_with("Line").len(), 3);
}