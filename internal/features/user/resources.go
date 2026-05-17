package user

type UserResource struct {
	User *User
}

func NewUserResource(user *User) UserResource {
	return UserResource{User: user}
}

func (r UserResource) ToResponse() SafeUser {
	if r.User == nil {
		return SafeUser{}
	}
	return r.User.ToSafeUser()
}

func NewUserCollection(users []User) []SafeUser {
	items := make([]SafeUser, len(users))
	for i := range users {
		items[i] = users[i].ToSafeUser()
	}
	return items
}
