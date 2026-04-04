package internal

type MapInfo struct {
	PosX   int     `json:"pos_x"`
	PosY   int     `json:"pos_y"`
	PixelX int     `json:"pixel_x"`
	PixelY int     `json:"pixel_y"`
	Scale  float32 `json:"scale"`
	Rotate float32 `json:"rotate"`
	Zoom   float32 `json:"zoom"`
}
