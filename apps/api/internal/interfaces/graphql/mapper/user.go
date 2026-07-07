package mapper

import (
	"time"

	"github.com/JosePintos/tora/apps/api/internal/interfaces/graphql/model"
	"github.com/JosePintos/tora/apps/api/internal/user/domain"
)

func ToGraphQLUser(user *domain.User) *model.User {
	return &model.User{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
}
