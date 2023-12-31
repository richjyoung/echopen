package main

type Order struct {
	ID       int64  `json:"id,omitempty"`
	PetID    int64  `json:"petId,omitempty"`
	Quantity int32  `json:"quantity,omitempty"`
	ShipDate string `json:"shipDate,omitempty"`
	Status   string `json:"status,omitempty" enum:"placed,approved,delivered" description:"Order Status"`
	Complete bool   `json:"complete,omitempty" default:"false"`
}

type Category struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type User struct {
	ID         int64  `json:"id,omitempty"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Email      string `json:"email,omitempty"`
	Password   string `json:"password,omitempty"`
	Phone      string `json:"phone,omitempty"`
	UserStatus int32  `json:"userStatus,omitempty" description:"User Status"`
}

type Tag struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Pet struct {
	ID        int64    `json:"id,omitempty"`
	Category  Category `json:"category,omitempty"`
	Name      string   `json:"name" example:"doggie"`
	PhotoURLs []string `json:"photoUrls"`
	Tags      []Tag    `json:"tags,omitempty"`
	Status    string   `json:"status,omitempty" enum:"available,pending,sold" description:"pet status in the store"`
}

type ApiResponse struct {
	Code    int32  `json:"code,omitempty"`
	Type    string `json:"type,omitempty"`
	Message string `json:"message,omitempty"`
}

type FindPetsByStatusQuery struct {
	Status []string `query:"status" description:"Status values that need to be considered for filter" enum:"available,pending,sold"`
}

type FindPetsByTagsQuery struct {
	Tags []string `query:"tags" description:"Tags to filter by"`
}

type UpdatePet struct {
	Name   string `json:"name,omitempty" description:"Updated name of the pet"`
	Status string `json:"status,omitempty" description:"Updated status of the pet"`
}

type GetOrderByIDPath struct {
	OrderID int64 `path:"orderId" description:"ID of order that needs to be fetched" validate:"min=1,max=10"`
}

type DeleteOrderByIDPath struct {
	OrderID int64 `path:"orderId" description:"ID of the order that needs to be deleted" validate:"min=1"`
}

type LoginUserQuery struct {
	Username string `query:"username" description:"The user name for login"`
	Password string `query:"password" description:"The password for login in clear text"`
}
