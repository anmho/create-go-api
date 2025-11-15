//go:build ignore

package posts

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	postsv1 "github.com/acme/postservice/internal/protos/gen/posts/v1"
)

// PostToProto converts a Post to a protobuf Post
func PostToProto(post *Post) *postsv1.Post {
	return &postsv1.Post{
		Id:        post.ID.String(),
		UserId:    post.UserID.String(),
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: timestamppb.New(post.CreatedAt),
		UpdatedAt: timestamppb.New(post.UpdatedAt),
	}
}
