package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
)

// Define a struct that matches the frontend JSON
type ContractRequest struct {
	BlockHeight int     `json:"blockHeight"`
	Price       float64 `json:"price"`
}

const simplicityTemplate = `/*
 * HODL VAULT
 * (auto-generated)
 */

fn checksig(pk: Pubkey, sig: Signature) {
    let msg: u256 = jet::sig_all_hash();
    jet::bip_0340_verify((pk, msg), sig);
}

fn checksigfromstack(pk: Pubkey, bytes: [u32; 2], sig: Signature) {
    let [word1, word2]: [u32; 2] = bytes;
    let hasher: Ctx8 = jet::sha_256_ctx_8_init();
    let hasher: Ctx8 = jet::sha_256_ctx_8_add_4(hasher, word1);
    let hasher: Ctx8 = jet::sha_256_ctx_8_add_4(hasher, word2);
    let msg: u256 = jet::sha_256_ctx_8_finalize(hasher);
    jet::bip_0340_verify((pk, msg), sig);
}

fn main() {
    let min_height: Height = %d;
    let oracle_height: Height = witness::ORACLE_HEIGHT;
    assert!(jet::le_32(min_height, oracle_height));
    jet::check_lock_height(oracle_height);

    let target_price: u32 = %d;
    let oracle_price: u32 = witness::ORACLE_PRICE;
    assert!(jet::le_32(target_price, oracle_price));

    let oracle_pk: Pubkey = 0x79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798; // 1 * G
    let oracle_sig: Signature = witness::ORACLE_SIG;
    checksigfromstack(oracle_pk, [oracle_height, oracle_price], oracle_sig);

    let owner_pk: Pubkey = 0xc6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5; // 2 * G
    let owner_sig: Signature = witness::OWNER_SIG;
    checksig(owner_pk, owner_sig);
}`

func main() {
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	app.Static("/", "./public", fiber.Static{
		Compress: true,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5174",
		AllowMethods: "GET,POST,OPTIONS",
	}))

	app.Post("/generate-contract", func(c *fiber.Ctx) error {
		var body ContractRequest

		// Parseing the json body into struct
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		fmt.Println("Received BlockHeight:", body.BlockHeight)
		fmt.Println("Received Price:", body.Price)

		// generating the simplicity code with dynamic values
		simpCode := fmt.Sprintf(simplicityTemplate, body.BlockHeight, int(body.Price))

		// saving into a temp file
		filePath := "./scripts/hodl_vault.simp"
		if err := os.WriteFile(filePath, []byte(simpCode), 0644); err != nil {
			return c.Status(500).SendString("Failed to write .simp file: " + err.Error())
		}

		// Run the compiler command
		//cmd := exec.Command("simc", "./scripts/hodl_vault.simp")
		cmd := exec.Command("simc", filePath)

		// the output
		output, err := cmd.CombinedOutput()
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error: %v\nOutput: %s", err, string(output)))
		}

		// Convert output to string and cleaning it
		cleanOutput := strings.TrimSpace(string(output))
		cleanOutput = strings.Replace(cleanOutput, "Program:", "", 1)
		cleanOutput = strings.TrimSpace(cleanOutput)

		// Run the bash script to get the address
		cmdscript := exec.Command("./scripts/get_addr.sh", cleanOutput)

		outputscript, err := cmdscript.CombinedOutput()
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error running script: %v\nOutput: %s", err, string(outputscript)))
		}

		// Clean result (the address)
		address := strings.TrimSpace(string(outputscript))

		return c.JSON(fiber.Map{
			"message":     "Contract generated successfully",
			"blockHeight": body.BlockHeight,
			"price":       body.Price,
			"totalValue":  float64(body.BlockHeight) * body.Price, // just an example
			"program":     cleanOutput,
			"address":     address,
		})
	})

	//This is just a test, a descontinued api route for now
	/*app.Post("/compile", func(c *fiber.Ctx) error {

		// Run the compiler command
		cmd := exec.Command("simc", "hodl_vault.simp")

		// Capture the output
		output, err := cmd.CombinedOutput() // captures stdout + stderr
		if err != nil {
			return c.Status(500).SendString(fmt.Sprintf("Error: %v\nOutput: %s", err, string(output)))
		}

		// Return the compiler output as JSON
		return c.JSON(fiber.Map{
			"result": string(output),
		})
	})*/

	log.Fatal(app.Listen(":4000"))
}
