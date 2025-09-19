#!/usr/bin/env python3
"""
CSV Newline Fixer

Fixes CSV files where quoted fields contain literal newlines by replacing
them with escaped newlines (\n) to maintain proper CSV structure.
"""

import sys
import argparse
import re
from pathlib import Path


def fix_csv_newlines(input_file: str, output_file: str = None) -> None:
    """
    Fix CSV files with newlines inside quoted fields.
    
    Args:
        input_file: Path to the input CSV file
        output_file: Path to the output CSV file (defaults to input_file if None)
    """
    input_path = Path(input_file)
    if not input_path.exists():
        raise FileNotFoundError(f"Input file not found: {input_file}")
    
    if output_file is None:
        output_file = input_file
    
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            content = f.read()
    except UnicodeDecodeError:
        # Fallback to other encodings if UTF-8 fails
        with open(input_file, 'r', encoding='latin-1') as f:
            content = f.read()
    
    # Fix newlines within quoted fields
    fixed_content = fix_quoted_newlines(content)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(fixed_content)
    
    print(f"Fixed CSV saved to: {output_file}")


def fix_quoted_newlines(content: str) -> str:
    """
    Replace literal newlines within quoted CSV fields with escaped newlines.
    
    This handles the case where CSV fields contain actual newline characters
    within double quotes, which breaks CSV parsing.
    """
    result = []
    in_quotes = False
    i = 0
    
    while i < len(content):
        char = content[i]
        
        if char == '"':
            # Handle escaped quotes ("")
            if i + 1 < len(content) and content[i + 1] == '"' and in_quotes:
                result.append('""')  # Keep escaped quotes as-is
                i += 2
                continue
            else:
                in_quotes = not in_quotes
                result.append(char)
        elif char == '\n' and in_quotes:
            # Replace newline with escaped newline when inside quotes
            result.append('\\n')
        elif char == '\r' and in_quotes:
            # Also handle carriage returns
            result.append('\\r')
        else:
            result.append(char)
        
        i += 1
    
    return ''.join(result)


def validate_csv_structure(file_path: str) -> bool:
    """
    Basic validation to check if the CSV structure looks reasonable.
    Returns True if the file appears to be valid CSV.
    """
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        if not lines:
            return False
        
        # Check if first line (header) has consistent comma count
        header_commas = lines[0].count(',')
        if header_commas == 0:
            return False
        
        # Check a few more lines to see if comma count is roughly consistent
        sample_size = min(5, len(lines))
        for i in range(1, sample_size):
            line_commas = lines[i].count(',')
            # Allow some variation but not too much
            if abs(line_commas - header_commas) > header_commas * 0.5:
                return False
        
        return True
    except Exception:
        return False


def main():
    parser = argparse.ArgumentParser(
        description="Fix CSV files with newlines in quoted fields",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s input.csv                    # Fix in place
  %(prog)s input.csv -o output.csv      # Save to different file
  %(prog)s input.csv --validate         # Fix and validate result
        """
    )
    
    parser.add_argument('input_file', help='Input CSV file to fix')
    parser.add_argument('-o', '--output', help='Output file (default: overwrite input)')
    parser.add_argument('--validate', action='store_true', 
                       help='Validate CSV structure after fixing')
    parser.add_argument('--backup', action='store_true',
                       help='Create backup of original file')
    
    args = parser.parse_args()
    
    try:
        # Create backup if requested
        if args.backup:
            backup_path = f"{args.input_file}.backup"
            Path(args.input_file).rename(backup_path)
            print(f"Backup created: {backup_path}")
            # Use backup as input, original name as output
            input_file = backup_path
            output_file = args.output or args.input_file
        else:
            input_file = args.input_file
            output_file = args.output
        
        # Fix the CSV
        fix_csv_newlines(input_file, output_file)
        
        # Validate if requested
        if args.validate:
            final_file = output_file or input_file
            if validate_csv_structure(final_file):
                print("✓ CSV structure validation passed")
            else:
                print("⚠ CSV structure validation failed - manual review recommended")
                sys.exit(1)
        
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()