package main

import (
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	tmpl *template.Template
}

func newTemplate() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

type Count struct {
	Count int
}

var id = Count{}

type Contact struct {
	Name  string
	Email string
	Id    int
}

type Contacts = []Contact

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func blankFormData() FormData {
	return FormData{
		Values: map[string]string{},
		Errors: map[string]string{},
	}
}

type PageData struct {
	Contacts Contacts
	FormData FormData
}

func newContact(name, email string) Contact {
	id.Count++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id.Count,
	}
}
func getContacts() Contacts {
	return Contacts{
		newContact("a", "a@a.com"),
		newContact("b", "b@b.com"),
		newContact("c", "c@c.com"),
	}
}

func getPageData() PageData {
	return PageData{
		Contacts: getContacts(),
		FormData: blankFormData(),
	}
}

func (d *PageData) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

func indexOf(contacts Contacts, id int) int {
	for i, c := range contacts {
		if c.Id == id {
			return i
		}
	}
	return -1
}

func main() {
	e := echo.New()
	e.Renderer = newTemplate()
	e.Use(middleware.Logger())
	e.Static("/images", "images")
	e.Static("/css", "css")

	count := Count{Count: 0}
	e.GET("/counter", func(c echo.Context) error {
		return c.Render(200, "counter", count)
	})
	e.POST("/count", func(c echo.Context) error {
		count.Count++
		return c.Render(200, "count", count)
	})

	pageData := getPageData()
	e.GET("/form", func(c echo.Context) error {
		return c.Render(200, "form-body", pageData)
	})
	e.POST("/contacts", func(c echo.Context) error {
		name := c.FormValue("name")
		email := c.FormValue("email")
		if pageData.hasEmail(email) {
			pageData.FormData.Values["name"] = name
			pageData.FormData.Values["email"] = email
			pageData.FormData.Errors["email"] = "Email already exists"

			return c.Render(422, "form", pageData.FormData)
		}
		c.Render(200, "form", blankFormData())

		contact := Contact{Name: name, Email: email}
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(2 * time.Second)
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid Id")
		}

		ind := indexOf(pageData.Contacts, id)
		if ind == -1 {
			return c.String(400, "Contact not found")
		}
		pageData.Contacts = append(pageData.Contacts[:ind], pageData.Contacts[ind+1:]...)
		return c.NoContent(200)
	})
	e.Logger.Fatal(e.Start(":42069"))
}
