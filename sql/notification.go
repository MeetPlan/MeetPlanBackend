package sql

type NotificationSQL struct {
	ID           string
	Notification string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetNotification(id string) (notification NotificationSQL, err error) {
	err = db.db.Get(&notification, "SELECT * FROM notifications WHERE id=$1", id)
	return notification, err
}

func (db *sqlImpl) GetAllNotifications() (notifications []NotificationSQL, err error) {
	err = db.db.Select(&notifications, "SELECT * FROM notifications")
	return notifications, err
}

func (db *sqlImpl) InsertNotification(notification NotificationSQL) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO notifications (notification) VALUES (:notification)",
		notification)
	return err
}

func (db *sqlImpl) UpdateNotification(notification NotificationSQL) error {
	_, err := db.db.NamedExec(
		"UPDATE notifications SET notification=:notification WHERE id=:id",
		notification)
	return err
}

func (db *sqlImpl) DeleteNotification(ID string) error {
	_, err := db.db.Exec("DELETE FROM notifications WHERE id=$1", ID)
	return err
}
