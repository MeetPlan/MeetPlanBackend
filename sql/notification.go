package sql

type NotificationSQL struct {
	ID           int
	Notification string

	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

func (db *sqlImpl) GetNotification(id int) (notification NotificationSQL, err error) {
	err = db.db.Get(&notification, "SELECT * FROM notifications WHERE id=$1", id)
	return notification, err
}

func (db *sqlImpl) GetAllNotifications() (notifications []NotificationSQL, err error) {
	err = db.db.Select(&notifications, "SELECT * FROM notifications")
	return notifications, err
}

func (db *sqlImpl) InsertNotification(notification NotificationSQL) (err error) {
	_, err = db.db.NamedExec(
		"INSERT INTO notifications (id, notification) VALUES (:id, :notification)",
		notification)
	return err
}

func (db *sqlImpl) UpdateNotification(notification NotificationSQL) error {
	_, err := db.db.NamedExec(
		"UPDATE notifications SET notification=:notification WHERE id=:id",
		notification)
	return err
}

func (db *sqlImpl) GetLastNotificationID() (id int) {
	err := db.db.Get(&id, "SELECT id FROM notifications WHERE id = (SELECT MAX(id) FROM notifications)")
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return 0
		}
		db.logger.Info(err)
		return -1
	}
	return id + 1
}

func (db *sqlImpl) DeleteNotification(ID int) error {
	_, err := db.db.Exec("DELETE FROM notifications WHERE id=$1", ID)
	return err
}
