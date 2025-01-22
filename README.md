# DirSpy

DirSpy is a tool that helps you find directories and files on a website.

## Parameters

- `-u`: **(required)** The base URL to crawl. For example: `http://example.com/`.
  
- `-i`: A comma-separated list of HTTP status codes to ignore during the crawl. For example: `404,403,500`.

- `-k`: A comma-separated list of keywords to search for within files. For example: `password,secret,key`.

- `-e`: A comma-separated list of file extensions to ignore. For example: `.txt,.jpg`.

- `-n`: Disable colored output. This is useful for logging purposes where color codes may not be supported.

- `-p`: Specify a proxy URL to use for the requests. Default is `http://localhost:8080`. For example: `http://proxyserver:8080`.

## Examples

1. Basic usage:
   ```bash
   ./dirspy -u=http://example.com/
   ```

2. Ignoring specific status codes:
   ```bash
   ./dirspy -u=http://example.com/ -i='404,403'
   ```

3. Searching for keywords:
   ```bash
   ./dirspy -u=http://example.com/ -k='password,api_key'
   ```

4. Ignoring file extensions:
   ```bash
   ./dirspy -u=http://example.com/ -e='.txt,.jpg'
   ```

5. Using a proxy:
   ```bash
   ./dirspy -u=http://example.com/ -p='http://localhost:8080'
   ```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any suggestions or improvements.

## Acknowledgments

- Thanks to the AI tools(Claude 3.5 Sonnet, ChatGPT 4o) for helping me with the code.
