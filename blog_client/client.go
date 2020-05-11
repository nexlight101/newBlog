package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nexlight101/gRPC_course/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func doReadBlog(c blogpb.BlogServiceClient) {
	fmt.Println("Reading a Blog")
	req := &blogpb.ReadBlogRequest{
		BlogId: "5eb9309c7ade23cccb8c797f",
	}

	res, sErr := c.ReadBlog(context.Background(), req)
	if sErr != nil {
		log.Fatal(status.Errorf(codes.InvalidArgument, fmt.Sprintf("Readblog gRPC Error: %v\n", sErr)))
	}
	fmt.Printf("We have found your blog %v\n", res.GetBlog())
}

func doUnary(c blogpb.BlogServiceClient) {
	fmt.Println("Sending the Blog request to server")
	blog := &blogpb.Blog{
		AuthorId: "Hennie",
		Title:    "My First Blog",
		Content:  "The sunshine on my shoulders make me happy!",
	}
	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v\n", err)
		return
	}
	fmt.Println("Blog has been created:")
	fmt.Println(res.GetBlog())
}

func main() {
	fmt.Println("Blog client")

	// Create connection to the server
	options := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", options)
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}

	// CLose the connection at exit
	defer cc.Close()

	// Establish a new client
	c := blogpb.NewBlogServiceClient(cc)
	fmt.Printf("Client activated: %v\n", c)

	// send request to unary client
	doUnary(c)

	doReadBlog(c)

}
