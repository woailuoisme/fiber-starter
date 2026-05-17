package resources

import (
	models "fiber-starter/app/Models"
)

type UserResource struct {
	User *models.User
}

func NewUserResource(user *models.User) UserResource {
	return UserResource{User: user}
}

func (r UserResource) ToResponse() models.SafeUser {
	if r.User == nil {
		return models.SafeUser{}
	}
	return r.User.ToSafeUser()
}

func NewUserCollection(users []models.User) []models.SafeUser {
	items := make([]models.SafeUser, len(users))
	for i := range users {
		items[i] = users[i].ToSafeUser()
	}
	return items
}
