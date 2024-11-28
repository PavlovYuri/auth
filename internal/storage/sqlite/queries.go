package sqlite

const (
	createUserQuery = `
		INSERT INTO users(email, pass_hash) VALUES(?, ?)
	`

	getUserQuery = `
		SELECT id, email, pass_hash 
		FROM users 
		WHERE email = ?
	`

	getAppQuery = `
		SELECT id, name, secret
		FROM apps
		WHERE id = ?
	`

	isAdminQuery = `
		SELECT is_admin
		FROM roles
		WHERE id = ?
	`
)
