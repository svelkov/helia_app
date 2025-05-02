package domain

type Type int

const (
	confirmModalType Type = iota
	infoModalType
)

// modal.go
type ModalDetail struct {
	Type          Type
	Heading       string
	Text          string
	CancelText    string
	ConfirmText   string
	ConfirmMethod string
	ConfirmAction string
}

// modal.go
func NewModalDetail(
	modalType Type,
	heading string,
	text string,
	cancelText string,
	confirmText string,
	confirmMethod string,
	confirmAction string,
) ModalDetail {
	if cancelText == "" {
		cancelText = "Cancel" // Default cancel button text
	}
	if confirmText == "" {
		confirmText = "OK" // Default confirm button text
	}
	if confirmMethod == "" {
		confirmMethod = "hx-post" // Default confirm method
	}
	return ModalDetail{
		Type:          modalType,
		Heading:       heading,
		Text:          text,
		CancelText:    cancelText,
		ConfirmText:   confirmText,
		ConfirmMethod: confirmMethod,
		ConfirmAction: confirmAction,
	}
}

// Defines the fields for creating a confirmation modal
type ConfirmModalState struct {
	Heading       string
	Text          string
	CancelText    string
	ConfirmMethod string
	ConfirmAction string
	ConfirmText   string
}

// NewConfirmModal creates a new confirmation modal
func NewConfirmModal(s ConfirmModalState) ModalDetail {
	return NewModalDetail(
		confirmModalType,
		s.Heading,
		s.Text,
		s.CancelText,
		s.ConfirmText,
		s.ConfirmMethod,
		s.ConfirmAction,
	)
}

// Define the fields required for the info modal
type InfoModalState struct {
	Heading     string
	Text        string
	ConfirmText string
}

// Create a new function for creating the modal and return a Render interface
func NewInfoModal(s InfoModalState) ModalDetail {
	return NewModalDetail(
		infoModalType, // Specify the new modal type
		s.Heading,
		s.Text,
		"", // Some of these params are not needed for this modal type
		s.ConfirmText,
		"",
		"",
	)
}
