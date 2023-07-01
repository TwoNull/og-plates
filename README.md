<br />
<div align="center">
  <a href="https://github.com/twonull/og-plates">
    <img src="https://raw.githubusercontent.com/TwoNull/og-plates/main/icon.png" alt="" width="160" height="80">
  </a>

<h3 align="center">OG Plates</h3>
  <p align="center">
    A utility to check bulk availability of Virginia DMV vanity plates.
    <br />
    <br />
    <a href="https://nullptrs.co/" target="_blank" rel="noopener noreferrer"><strong>Read the Full Writeup »</strong></a>
    <br />
  </p>
</div>


### Disclaimer

This tool is meant to be a proof-of-concept only. I take no responsibility for the unintended consequences of sending thousands of simultaneous requests to a government website.

<br />

### Usage

1. Download the latest [release](https://github.com/TwoNull/og-plates/releases/tag/Major)
   <br />
   <br />
2. Unzip and run the tool from the command line:
   \
   MacOS/Linux
   ```sh
   $ ./og-plates [args]
   ```
   Windows
   ```bash
   $ og-plates.exe [args]
   ```
   <br />
3. A list of arguments will be output to the command line.

<p align="right">(<a href="#readme-top">back to top</a>)</p>



### Made With

[![Golang][Golang]][Go-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Build it yourself

Below are the prerequisites to build your own version of this tool from source.
* [Go ^1.20](https://go.dev/dl/)

### Setup

1. Clone the repo
   ```sh
   $ git clone https://github.com/TwoNull/og-plates.git
   ```
   <br />
2. Navigate to the project directory and build the tool
   ```sh
   $ cd og-plates
   $ go build
   ```
   <br />
3. The executable will output as `og-plates` or `og-plates.exe` depending on your environment.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>


<!-- MARKDOWN LINKS & IMAGES -->
[Golang]: https://shields.io/badge/Golang-5DC9E2?style=for-the-badge&logo=Go&logoColor=FFF
[Go-url]: https://go.dev/
