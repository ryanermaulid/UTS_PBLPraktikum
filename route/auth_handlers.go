package route

import (
	"github.com/gofiber/fiber/v2"

	"github.com/ryanermaulid/UTS_PBLPraktikum/app/model"
	"github.com/ryanermaulid/UTS_PBLPraktikum/app/service"
	"github.com/ryanermaulid/UTS_PBLPraktikum/helper"
)

// loginHandler mem-parsing body, memvalidasi via helper.ValidateStruct,
// lalu mendelegasikan ke AuthService.Login. Validasi menghasilkan 422
// dengan FormatErrorService, sedangkan kredensial salah menghasilkan 401
// dari service.
func loginHandler(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req model.LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return helper.Unprocessable("Body tidak valid")
		}
		if err := helper.ValidateStruct(&req); err != nil {
			return err
		}
		resp, err := svc.Login(c.UserContext(), req)
		if err != nil {
			return err
		}
		return helper.OK(c, "Login berhasil", resp)
	}
}

// meHandler mengembalikan data user yang sedang login berdasarkan
// userID yang disimpan RequireAuth di locals.
func meHandler(svc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		v := c.Locals("user_id")
		id, ok := v.(int64)
		if !ok {
			return helper.Unauthorized("Token tidak ditemukan")
		}
		resp, err := svc.GetMe(c.UserContext(), id)
		if err != nil {
			return err
		}
		return helper.OK(c, "Data user saat ini", resp)
	}
}
