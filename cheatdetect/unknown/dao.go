package unknown

import "encoding/xml"

type UsersResponse struct {
	XMLName xml.Name `xml:"users"`
	Users   []User   `xml:"user"`
}

type User struct {
	XMLName  xml.Name `xml:"user"`
	UserID   string   `xml:"userid,attr"`
	Username string   `xml:",chardata"`
}
