package init_cmd

import (
	"flag"
	"fmt"
	"os"
)

const template = `# specter config
# docs: https://github.com/Saku0512/specter

routes:
  - path: /hello
    method: GET
    response:
      message: Hello, World!

  - path: /users
    method: GET
    response:
      - id: 1
        name: Alice
      - id: 2
        name: Bob

  - path: /users/:id
    method: GET
    response:
      id: ":id"
      name: Alice

  - path: /users
    method: POST
    status: 201
    response:
      message: created
`

func Run(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("o", "config.yml", "output file")
	force := fs.Bool("f", false, "overwrite if file already exists")
	fs.Parse(args)

	if _, err := os.Stat(*output); err == nil && !*force {
		fmt.Fprintf(os.Stderr, "%s already exists. Use -f to overwrite.\n", *output)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, []byte(template), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", *output, err)
		os.Exit(1)
	}

	fmt.Printf("created %s\n", *output)
	fmt.Println("run: specter -c", *output)
}
