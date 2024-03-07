package entity

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterResponse struct {
	Id int `json:"id"`
}

type UserAuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuthResponse struct {
	Token string `json:"token"`
}

type CategoryCreateRequest struct {
	CategoryName string `json:"category_name"`
}

type CategoryCreateResponse struct {
	CategoryId   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

type CategoryEditRequest struct {
	CategoryId int    `json:"category_id"`
	NewName    string `json:"new_name"`
}

type CategoryEditResponse struct {
	CategoryId int    `json:"category_id"`
	NewName    string `json:"new_name"`
}

type CategoryDeleteResponse struct {
	CategoryId int  `json:"category_id"`
	Deleted    bool `json:"deleted"`
}

type GoodAddRequest struct {
	GoodName   string `json:"good_name"`
	CategoryId int    `json:"category_id"`
}

type GoodAddResponse struct {
	GoodId         int    `json:"good_id"`
	GoodCategoryId int    `json:"good_category_id"`
	GoodName       string `json:"good_name"`
	CategoryName   string `json:"category_name"`
}

type GoodUpdateRequest struct {
	GoodId          int    `json:"good_id"`
	GoodActualName  string `json:"good_actual_name,omitempty"`
	AddedCategoryId int    `json:"added_category_id,omitempty"`
}

type GoodUpdateResponse struct {
	GoodId       int      `json:"good_id"`
	GoodName     string   `json:"good_name"`
	CategoryName []string `json:"category_name"`
}

type GoodDeleteResponse struct {
	GoodId  int  `json:"good_id"`
	Deleted bool `json:"deleted"`
}

type CategoryList struct {
	CategoryId   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

type GoodList struct {
	GoodId   int    `json:"good_id"`
	GoodName string `json:"good_name"`
}
