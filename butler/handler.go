package butler

import (
	"net/http"
)

type HTTPHandleFunc func(b *Butler, w http.ResponseWriter, r *http.Request)
