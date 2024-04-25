# Dockerium

Dockerium is my own version of Docker that I made using Golang. I did this for two main reasons:

1. I wanted to learn Golang.
2. I wanted to understand how Docker works behind the scenes.

## Video Demo
[Screencast from 25-04-24 08:42:38 PM IST.webm](https://github.com/Vandit1604/dockerium/assets/107131545/29a32c59-183b-475c-81f2-908e5d41f990)


## Usage

To use Dockerium, follow these steps:

1. Build the binary:

```
make build
```

2. Run the project with the desired image name:

```
./dockerium <name of the image>
```

This will fetch the image layers and extract them for you to use.

## Running Locally

To run Dockerium locally, you need to have Golang installed on your system. Follow these steps:

1. Clone the repository:

```
git clone https://github.com/Vandit1604/dockerium.git
```

2. Navigate to the project directory:

```
cd dockerium
```

3. Build the binary:

```
make build
```

4. Run the project:

```
./dockerium <name of the image>
```

## Contributing

Contributions to Dockerium are welcome! To contribute, follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes.
4. Test your changes thoroughly.
5. Commit your changes with clear and descriptive messages.
6. Push your changes to your fork.
7. Open a pull request against the main repository.

Please ensure that your code adheres to the project's coding standards and practices.
