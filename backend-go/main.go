package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
)

// Define a struct that matches the frontend JSON
type ContractRequest struct {
	BlockHeight int     `json:"blockHeight"`
	Price       float64 `json:"price"`
	//OraclePubKey string  `json:"oracle_pubkey"`
}

type AddressRequest struct {
	Address string `json:"address"`
}

type FaucetResponse struct {
	TxID  string `json:"txId"`
	Error string `json:"error"`
}

const witnessTemplate = `
{
    "ORACLE_HEIGHT": {
        "value": "<BlockHeight>",
        "type": "u32"
    },
    "ORACLE_PRICE": {
        "value": "<Price>",
        "type": "u32"
    },
    "ORACLE_SIG": {
        "value": "<OracleSig>",
        "type": "Signature"
    }
}
`

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

    let oracle_pk: Pubkey = 0x79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798; //for the moment hardcoded
    let oracle_sig: Signature = witness::ORACLE_SIG;
    checksigfromstack(oracle_pk, [oracle_height, oracle_price], oracle_sig);

    /*
    * The owner signature check is gone.
    * If the script reaches this point, all oracle checks have passed,
    * and the transaction is considered valid.
    * let owner_pk: Pubkey = 0xc6047f9441ed7d6d3045406e95c07cd85c778e4b8cef3ca7abac09b95c709ee5; // 2 * G
    * let owner_sig: Signature = witness::OWNER_SIG;
    * checksig(owner_pk, owner_sig);
    */
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
		//fmt.Println("Oracle Pubkey:", body.OraclePubKey)

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
			"program_hex": cleanOutput,
			"address":     address,
		})
	})

	app.Post("/fund-contract", func(c *fiber.Ctx) error {
		var body AddressRequest
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON: " + err.Error()})
		}

		// 1. Construa a URL (como antes)
		faucetURL := "https://liquidtestnet.com/faucet"
		fullURL := fmt.Sprintf("%s?address=%s&action=lbtc", faucetURL, body.Address)

		// 2. Execute o GET (como antes)
		resp, err := http.Get(fullURL)
		if err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Faucet service is unavailable: " + err.Error()})
		}
		defer resp.Body.Close()

		// 3. Verifique o Status Code (como antes)
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return c.Status(resp.StatusCode).JSON(fiber.Map{
				"error":           "Faucet API returned a non-200 status.",
				"status_code":     resp.StatusCode,
				"faucet_response": string(bodyBytes),
			})
		}

		// 4. Leia o corpo HTML (como antes)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to read faucet response body: " + err.Error()})
		}

		// --- INÍCIO DA NOVA LÓGICA (SUBSTITUA O 'json.Unmarshal') ---

		bodyString := string(bodyBytes)

		// 5. Defina a RegEx para encontrar o TxID
		// A txid é uma string de 64 caracteres hexadecimais (a-f, 0-9)
		// A RegEx procura por: "with transaction " seguido por (64 chars hex) e ".</p>"
		re := regexp.MustCompile(`with transaction ([a-f0-9]{64})\.</p>`)

		// 6. Tente encontrar o padrão na resposta HTML
		matches := re.FindStringSubmatch(bodyString)

		// 7. Verifique se encontramos
		// matches[0] é o texto completo (ex: "with transaction ...</p>")
		// matches[1] é o primeiro grupo de captura (apenas o txId)
		if len(matches) < 2 {
			// Não encontrou o TxID!
			// Provavelmente a faucet retornou um erro 200 OK, mas com uma
			// mensagem de erro no HTML (ex: "Address already funded")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":           "Failed to find transaction ID in faucet HTML response.",
				"faucet_response": bodyString, // Envie o HTML para depuração
			})
		}

		// 8. Sucesso! Extraímos o TxID
		txid := matches[1]

		return c.JSON(fiber.Map{
			"txid": txid,
		})

		// --- FIM DA NOVA LÓGICA ---
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
