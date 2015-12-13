package db_test

import (
	"testing"

	db "github.com/fiatjaf/summadb/database"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestAuthUsersACL(t *testing.T) {
	g := Goblin(t)
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("users and permissions", func() {
		g.Before(func() {
			db.Erase()
			db.Start()
		})

		g.After(func() {
			db.End()
		})

		g.It("should start with a * permission on root", func() {
			Expect(db.GetWriteRuleAt("/")).To(Equal("*"))
			Expect(db.GetReadRuleAt("/")).To(Equal("*"))
			Expect(db.GetAdminRuleAt("")).To(Equal("*")) // "/" should equal ""
		})

		g.It("which means everybody is allowed to do anything anywhere", func() {
			Expect(db.WriteAllowedAt("/", "bob")).To(Equal(true))
			Expect(db.ReadAllowedAt("/articles", "maria")).To(Equal(true))
			Expect(db.AdminAllowedAt("/somewhere/down/on/the/path", "anyone")).To(Equal(true))
			Expect(db.WriteAllowedAt("/x/r/e/ws/sd/f/t/r/e/d/g/y", "")).To(Equal(true))
			Expect(db.ReadAllowedAt("/recipes", "anna")).To(Equal(true))
			Expect(db.AdminAllowedAt("/", "bob")).To(Equal(true))
		})

		g.It("should modify permissions arbitrarily", func() {
			Expect(db.SetRulesAt("/paper", map[string]interface{}{
				"_read":  "drawer,romancist, reader",
				"_write": "drawer, romancist",
				"_admin": "romancist",
			})).To(Succeed())
			Expect(db.SetRulesAt("/", map[string]interface{}{
				"_read":  "myself",
				"_write": "myself",
				"_admin": "myself",
			})).To(Succeed())
		})

		g.It("and allowability should reflect that", func() {
			Expect(db.WriteAllowedAt("/", "bob")).To(Equal(false))
			Expect(db.ReadAllowedAt("/articles", "maria")).To(Equal(false))
			Expect(db.AdminAllowedAt("/somewhere/down/on/the/path", "anyone")).To(Equal(false))
			Expect(db.ReadAllowedAt("/recipes", "anna")).To(Equal(false))
			Expect(db.WriteAllowedAt("/x/r/e/ws/sd/f/t/r/e/d/g/y", "")).To(Equal(false))
			Expect(db.AdminAllowedAt("/", "bob")).To(Equal(false))

			Expect(db.WriteAllowedAt("/", "myself")).To(Equal(true))
			Expect(db.WriteAllowedAt("/paper", "myself")).To(Equal(true))
			Expect(db.WriteAllowedAt("/rock", "myself")).To(Equal(true))
			Expect(db.WriteAllowedAt("/paper/planes", "myself")).To(Equal(true))
			Expect(db.WriteAllowedAt("/paper", "romancist")).To(Equal(true))
			Expect(db.WriteAllowedAt("/paper/planes", "romancist")).To(Equal(true))
			Expect(db.WriteAllowedAt("/", "romancist")).To(Equal(false))
			Expect(db.WriteAllowedAt("/paperless", "romancist")).To(Equal(false))

			Expect(db.ReadAllowedAt("/paper", "drawer")).To(Equal(true))
			Expect(db.ReadAllowedAt("/paperless", "drawer")).To(Equal(false))
			Expect(db.ReadAllowedAt("/paper/origami", "drawer")).To(Equal(true))
			Expect(db.ReadAllowedAt("/paper", "reader")).To(Equal(true))
			Expect(db.ReadAllowedAt("/paperless", "reader")).To(Equal(false))
			Expect(db.ReadAllowedAt("/paper/origami", "reader")).To(Equal(true))

			Expect(db.AdminAllowedAt("/", "myself")).To(Equal(true))
			Expect(db.AdminAllowedAt("/anywhere", "myself")).To(Equal(true))
			Expect(db.AdminAllowedAt("/paper", "myself")).To(Equal(true))
			Expect(db.AdminAllowedAt("/", "romancist")).To(Equal(false))
			Expect(db.AdminAllowedAt("/anywhere", "romancist")).To(Equal(false))
			Expect(db.AdminAllowedAt("/paper", "romancist")).To(Equal(true))
			Expect(db.AdminAllowedAt("/paper/origami", "romancist")).To(Equal(true))
			Expect(db.AdminAllowedAt("/", "reader")).To(Equal(false))
			Expect(db.AdminAllowedAt("/anywhere", "reader")).To(Equal(false))
			Expect(db.AdminAllowedAt("/paper", "reader")).To(Equal(false))
		})

		g.It("should create users", func() {
			Expect(db.SaveUser("isaiah", "12345678")).To(Succeed())
			Expect(db.SaveUser("samuel", "87654321")).To(Succeed())
			Expect(db.SaveUser("isaiah", "qwertyuu")).ToNot(Succeed())
			Expect(db.SaveUser("israel", "")).To(Succeed())
			Expect(db.SaveUser("", "asdfghjh")).ToNot(Succeed())
		})

		g.It("should validate user logins", func() {
			Expect(db.ValidUser("samuel", "87654321")).To(Equal(true))
			Expect(db.ValidUser("isaiah", "12345678")).To(Equal(true))
			Expect(db.ValidUser("isaiah", "qwertyuu")).To(Equal(false))
			Expect(db.ValidUser("israel", "")).To(Equal(true))
			Expect(db.ValidUser("", "asdfghjh")).To(Equal(false))

			Expect(db.ValidUser("", "")).To(Equal(false))
			Expect(db.ValidUser("q", "a")).To(Equal(false))
			Expect(db.ValidUser("weq", "")).To(Equal(false))
			Expect(db.ValidUser("", "ssdds")).To(Equal(false))
		})
	})
}
