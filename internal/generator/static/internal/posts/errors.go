package posts

import "errors"

var ErrPostNotFound error = errors.New("post not found")