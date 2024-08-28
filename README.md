# GOSKI


Goski is converts images into ASCII art.
It supports both colored and grayscale ASCII art and can automatically scale the output to fit within your terminal window.

## Installation

Recommended
```sh
go install github.com/musaubrian/goski@latest
```
or
```sh
git clone https://github.com/musaubrian/goski

cd goski
go build .

#run the executable
```

## Usage
Run the program with the following command-line options:
```sh
goski -img <path-to-image> [-c] [-s]
```

Options:
- `img <image-path>`: Specifies the path to the image file you want to convert.**(Required)**
- `c`: Outputs ASCII art with color. By default, the ASCII art is grayscale.
- `s`: Automatically *scales the output to fit the dimensions of your terminal.
If not specified, the image will be converted to ASCII art at its original size.
