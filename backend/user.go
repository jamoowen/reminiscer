
type User struct {
	ID             int    `json:"id"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	HashedPassword string `json:"hashed_password"`
	Authenticated  bool   `json:"authenticated"`
}
type PersistedUser struct {
	User      `bson:",inline" json:",inline"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
