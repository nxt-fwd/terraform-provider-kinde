package provider

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource_ComplexAttributes(t *testing.T) {
	testID := rand.Int()
	email := fmt.Sprintf("complex.user.tfacc-%d@example.com", testID)
	altEmail := fmt.Sprintf("complex.user.alt.tfacc-%d@example.com", testID)
	username := fmt.Sprintf("complex-user-%d", testID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a user with email and username identities, is_suspended=false
			{
				Config: testAccUserResourceConfig_ComplexAttributes(email, username, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.complex", "first_name", "Complex"),
					resource.TestCheckResourceAttr("kinde_user.complex", "last_name", "User"),
					resource.TestCheckResourceAttr("kinde_user.complex", "is_suspended", "false"),
					resource.TestCheckResourceAttr("kinde_user.complex", "identities.#", "2"),
					// Check that both identities exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.complex", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.complex", "identities.*", map[string]string{
						"type":  "username",
						"value": username,
					}),
				),
			},
			// Update is_suspended to true and add another email identity
			{
				Config: testAccUserResourceConfig_ComplexAttributesWithAltEmail(email, altEmail, username, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.complex", "first_name", "Complex"),
					resource.TestCheckResourceAttr("kinde_user.complex", "last_name", "User"),
					resource.TestCheckResourceAttr("kinde_user.complex", "is_suspended", "true"),
					resource.TestCheckResourceAttr("kinde_user.complex", "identities.#", "3"),
					// Check that all identities exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.complex", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.complex", "identities.*", map[string]string{
						"type":  "email",
						"value": altEmail,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.complex", "identities.*", map[string]string{
						"type":  "username",
						"value": username,
					}),
				),
			},
		},
	})
}

func testAccUserResourceConfig_ComplexAttributes(email, username string, isSuspended bool) string {
	return fmt.Sprintf(`
resource "kinde_user" "complex" {
	first_name = "Complex"
	last_name = "User"
	is_suspended = %[3]t

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "username"
			value = %[2]q
		}
	]
}
`, email, username, isSuspended)
}

func testAccUserResourceConfig_ComplexAttributesWithAltEmail(email, altEmail, username string, isSuspended bool) string {
	return fmt.Sprintf(`
resource "kinde_user" "complex" {
	first_name = "Complex"
	last_name = "User"
	is_suspended = %[4]t

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "username"
			value = %[3]q
		},
		{
			type = "email"
			value = %[2]q
		}
	]
}
`, email, altEmail, username, isSuspended)
}

func TestAccUserResource_PhoneIdentity(t *testing.T) {
	testID := rand.Int()
	email := fmt.Sprintf("phone.user.tfacc-%d@example.com", testID)
	phone := "+61412345678"  // Australian format with country code
	altPhone := "+12025550123"  // US format with country code

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a user with email and phone identities
			{
				Config: testAccUserResourceConfig_WithPhone(email, phone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.phone", "first_name", "Phone"),
					resource.TestCheckResourceAttr("kinde_user.phone", "last_name", "User"),
					resource.TestCheckResourceAttr("kinde_user.phone", "identities.#", "2"),
					// Check that both identities exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.phone", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.phone", "identities.*", map[string]string{
						"type":  "phone",
						"value": phone,
					}),
				),
			},
			// Add another phone identity
			{
				Config: testAccUserResourceConfig_WithMultiplePhones(email, phone, altPhone),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.phone", "identities.#", "3"),
					// Check that all identities exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.phone", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.phone", "identities.*", map[string]string{
						"type":  "phone",
						"value": phone,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.phone", "identities.*", map[string]string{
						"type":  "phone",
						"value": altPhone,
					}),
				),
			},
		},
	})
}

func testAccUserResourceConfig_WithPhone(email, phone string) string {
	return fmt.Sprintf(`
resource "kinde_user" "phone" {
	first_name = "Phone"
	last_name = "User"

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "phone"
			value = %[2]q
		}
	]
}
`, email, phone)
}

func testAccUserResourceConfig_WithMultiplePhones(email, phone1, phone2 string) string {
	return fmt.Sprintf(`
resource "kinde_user" "phone" {
	first_name = "Phone"
	last_name = "User"

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "phone"
			value = %[2]q
		},
		{
			type = "phone"
			value = %[3]q
		}
	]
}
`, email, phone1, phone2)
}

func TestAccUserResource_OAuth2Identity(t *testing.T) {
	testID := rand.Int()
	email := fmt.Sprintf("oauth2.user.tfacc-%d@example.com", testID)
	username := fmt.Sprintf("oauth2-user-%d", testID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create a user with email and username identities
			{
				Config: testAccUserResourceConfig_OAuth2(email, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.oauth2", "first_name", "OAuth2"),
					resource.TestCheckResourceAttr("kinde_user.oauth2", "last_name", "User"),
					resource.TestCheckResourceAttr("kinde_user.oauth2", "identities.#", "2"),
					// Check that both identities exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.oauth2", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.oauth2", "identities.*", map[string]string{
						"type":  "username",
						"value": username,
					}),
				),
			},
			// Update user details while preserving identities
			{
				Config: testAccUserResourceConfig_OAuth2Updated(email, username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.oauth2", "first_name", "Updated"),
					resource.TestCheckResourceAttr("kinde_user.oauth2", "last_name", "OAuth2"),
					resource.TestCheckResourceAttr("kinde_user.oauth2", "identities.#", "2"),
					// Check that both identities still exist
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.oauth2", "identities.*", map[string]string{
						"type":  "email",
						"value": email,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("kinde_user.oauth2", "identities.*", map[string]string{
						"type":  "username",
						"value": username,
					}),
				),
			},
		},
	})
}

func testAccUserResourceConfig_OAuth2(email, username string) string {
	return fmt.Sprintf(`
resource "kinde_user" "oauth2" {
	first_name = "OAuth2"
	last_name = "User"

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "username"
			value = %[2]q
		}
	]
}
`, email, username)
}

func testAccUserResourceConfig_OAuth2Updated(email, username string) string {
	return fmt.Sprintf(`
resource "kinde_user" "oauth2" {
	first_name = "Updated"
	last_name = "OAuth2"

	identities = [
		{
			type = "email"
			value = %[1]q
		},
		{
			type = "username"
			value = %[2]q
		}
	]
}
`, email, username)
}

func TestAccUserResource_NameHandling(t *testing.T) {
	testID := rand.Int()
	email := fmt.Sprintf("name.test.tfacc-%d@example.com", testID)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with both names set
			{
				Config: testAccUserResourceConfig_Names(email, "John", "Doe"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.name_test", "first_name", "John"),
					resource.TestCheckResourceAttr("kinde_user.name_test", "last_name", "Doe"),
				),
			},
			// Update with new non-empty values
			{
				Config: testAccUserResourceConfig_Names(email, "Jane", "Smith"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.name_test", "first_name", "Jane"),
					resource.TestCheckResourceAttr("kinde_user.name_test", "last_name", "Smith"),
				),
			},
			// Omit fields - should preserve previous values
			{
				Config: testAccUserResourceConfig_NamesOmitted(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("kinde_user.name_test", "first_name"),
					resource.TestCheckNoResourceAttr("kinde_user.name_test", "last_name"),
				),
			},
			// Set names again to verify we can still update after omitting
			{
				Config: testAccUserResourceConfig_Names(email, "Alice", "Johnson"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kinde_user.name_test", "first_name", "Alice"),
					resource.TestCheckResourceAttr("kinde_user.name_test", "last_name", "Johnson"),
				),
			},
		},
	})
}

func testAccUserResourceConfig_Names(email, firstName, lastName string) string {
	return fmt.Sprintf(`
resource "kinde_user" "name_test" {
	first_name = %[2]q
	last_name = %[3]q
	identities = [
		{
			type = "email"
			value = %[1]q
		}
	]
}
`, email, firstName, lastName)
}

func testAccUserResourceConfig_NamesOmitted(email string) string {
	return fmt.Sprintf(`
resource "kinde_user" "name_test" {
	identities = [
		{
			type = "email"
			value = %[1]q
		}
	]
}
`, email)
} 