package user

type User struct {
	ID   int    `db:"id" json:"json_id"`
	Name string `db:"nickname" json:"json_name"`
}

type Person struct {
	Fname string `db:"first_name" json:"fname"`
}
