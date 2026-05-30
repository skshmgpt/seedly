package prisma

import "testing"

func TestParseModelWithRelation(t *testing.T) {
	schema, err := ParseString(`
model User {
  id    Int    @id @default(autoincrement())
  email String @unique
  name  String?
  orders Order[]
}

model Order {
  id     Int  @id @default(autoincrement())
  userId Int
  user   User @relation(fields: [userId], references: [id])
  total  Decimal
}
`)
	if err != nil {
		t.Fatalf("ParseString returned error: %v", err)
	}

	order, ok := schema.Model("Order")
	if !ok {
		t.Fatal("expected Order model")
	}

	userField, ok := order.Field("user")
	if !ok {
		t.Fatal("expected user relation field")
	}
	if userField.Relation == nil {
		t.Fatal("expected relation metadata")
	}
	if userField.Relation.TargetModel != "User" {
		t.Fatalf("target model = %q, want User", userField.Relation.TargetModel)
	}
	if got := userField.Relation.Fields[0]; got != "userId" {
		t.Fatalf("relation field = %q, want userId", got)
	}
}

func TestParseRejectsMalformedModel(t *testing.T) {
	_, err := ParseString(`model User { id Int`)
	if err == nil {
		t.Fatal("expected error")
	}
}
