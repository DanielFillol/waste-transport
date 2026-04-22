package entity

type Material struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
}

type Packaging struct {
	ID     uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name   string  `gorm:"not null" json:"name"`
	Type   string  `json:"type"`
	Volume float64 `json:"volume"`
}

type Treatment struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
}

type UF struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null" json:"name"`
	Code string `gorm:"size:2;not null" json:"code"`
}

type City struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null" json:"name"`
	UFID uint   `gorm:"not null" json:"uf_id"`
	UF   *UF    `gorm:"foreignKey:UFID" json:"uf,omitempty"`
}
