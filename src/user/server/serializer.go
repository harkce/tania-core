package server

import (
	"github.com/Tanibox/tania-server/src/user/domain"
	"github.com/Tanibox/tania-server/src/user/storage"
)

func MapToUserRead(user *domain.User) storage.UserRead {
	userRead := storage.UserRead{}
	userRead.UID = user.UID
	userRead.Username = user.Username
	userRead.CreatedDate = user.CreatedDate
	userRead.LastUpdated = user.LastUpdated

	return userRead
}
