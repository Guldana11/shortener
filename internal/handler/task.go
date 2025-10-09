package handler

//
//import (
//	"errors"
//	"math"
//)
//
//func Abs(value float64) float64 {
//	return math.Abs(value)
//}
//
////type User struct {
////	FirstName string
////	LastName  string
////}
//
//func (u User) FullName() string {
//	return u.FirstName + " " + u.LastName
//}
//
//type Relationship string
//
//const (
//	Father      = Relationship("father")
//	Mother      = Relationship("mother")
//	Child       = Relationship("child")
//	GrandMother = Relationship("grandMother")
//	GrandFather = Relationship("grandFather")
//)
//
//type Person struct {
//	FirstName string
//	LastName  string
//	Age       int
//}
//
//type Family struct {
//	Members map[Relationship]Person
//}
//
//var (
//	ErrRelationshipAlreadyExists = errors.New("relationship already exists")
//)
//
//func (f *Family) AddNew(r Relationship, p Person) error {
//	if f.Members == nil {
//		f.Members = map[Relationship]Person{}
//	}
//	if _, ok := f.Members[r]; ok {
//		return ErrRelationshipAlreadyExists
//	}
//	f.Members[r] = p
//	return nil
//}
//
////func main() {
////	// ABS
////	v := Abs(-3.14)
////	fmt.Println("Abs(-3.14) =", v)
////
////	// USER
////	u := User{
////		FirstName: "Misha",
////		LastName:  "Popov",
////	}
////	fmt.Println("User FullName:", u.FullName())
////
////	// FAMILY
////	f := Family{}
////	err := f.AddNew(Father, Person{
////		FirstName: "Misha",
////		LastName:  "Popov",
////		Age:       56,
////	})
////	fmt.Println("Family after adding father:", f.Members, "Error:", err)
////
////	err = f.AddNew(Father, Person{
////		FirstName: "Drug",
////		LastName:  "Mishi",
////		Age:       57,
////	})
////	fmt.Println("Family after adding second father:", f.Members, "Error:", err)
////}
