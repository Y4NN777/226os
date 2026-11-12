package httpx

import "github.com/Y4NN777/226os/src/api/internal/platform/apperr"

// StatusTable exposes the single code->status table to the contract test.
func StatusTable() map[apperr.Code]int { return statusTable }

// TitleTable exposes the code->title table to the contract test.
func TitleTable() map[apperr.Code]string { return titleTable }
