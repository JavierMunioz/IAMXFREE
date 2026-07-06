// Package dashboard is the main screen of IAMXFREE: a card grid showing
// every registered application, a top status bar, and a bottom keybinding
// bar. It is an independent presentation component — it consumes only
// services.ApplicationService and never touches a repository directly, and
// it holds no business logic of its own (validation, persistence, and
// conflict handling all live in the service layer).
package dashboard
