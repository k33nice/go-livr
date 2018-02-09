// Package livr is Language Independent Validation Rules Specification (LIVR) for Go.
// Basically for using in web services for validation JSON data.
//
// Features:
//   * Rules are declarative and language independent
//   * Any number of rules for each field
//   * Return together errors for all fields
//   * Excludes all fields that do not have validation rules described
//   * Has possibility to validate complex hierarchical structures
//   * Easy to describe and understand rules
//   * Returns understandable error codes(not error messages)
//   * Easy to add own rules
//   * Rules are be able to change results output ("trim", "nested_object", for example)
//   * Multipurpose (user input validation, configs validation, contracts programming etc)
package livr
