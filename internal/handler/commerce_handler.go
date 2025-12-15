package handler

import (
	"fmt"
	"path/filepath"
	"umrah-backend/internal/entity"
	"umrah-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CommerceHandler struct {
	svc service.CommerceService
}

func NewCommerceHandler(svc service.CommerceService) *CommerceHandler {
	return &CommerceHandler{svc: svc}
}

// POST /products (Admin Only)
func (h *CommerceHandler) CreateProduct(c *fiber.Ctx) error {
	var req entity.Product
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := h.svc.CreateProduct(c.Context(), req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "Product created"})
}

// GET /products (Catalog)
func (h *CommerceHandler) GetCatalog(c *fiber.Ctx) error {
	products, err := h.svc.GetCatalog(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(products)
}

// POST /orders (Buy)
func (h *CommerceHandler) CreateOrder(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	var req struct {
		ProductID string `json:"product_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	order, err := h.svc.CreateOrder(c.Context(), userID, req.ProductID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(order)
}

// GET /orders/my (History)
func (h *CommerceHandler) GetMyOrders(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	orders, err := h.svc.GetMyOrders(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

// POST /orders/:id/proof (Upload Image)
func (h *CommerceHandler) UploadProof(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	orderID := c.Params("id")

	// 1. Handle File Upload
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Image required"})
	}

	// 2. Save File Locally (In Production: Use S3)
	// Make sure 'uploads' folder exists in your project root!
	filename := fmt.Sprintf("%s_%s%s", userID, uuid.New().String(), filepath.Ext(file.Filename))
	savePath := fmt.Sprintf("./uploads/%s", filename)

	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save image"})
	}

	// 3. Call Service
	// We store the relative path or full URL
	publicURL := fmt.Sprintf("/uploads/%s", filename)
	if err := h.svc.UploadPaymentProof(c.Context(), orderID, publicURL, userID); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Proof uploaded", "url": publicURL})
}

// PATCH /orders/:id/verify (Admin Only)
func (h *CommerceHandler) VerifyOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if err := h.svc.VerifyOrder(c.Context(), orderID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "Order verified"})
}
