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

func doDeleteBlog(c blogpb.BlogServiceClient) {
	fmt.Println("Deleting a Blog")
	// populate a requets
	req := &blogpb.DeleteBlogRequest{
		BlogId: "5e97033cfd1bae8eadfccf24",
	}
	// Sends a request
	res, dErr := c.DeleteBlog(context.Background(), req)
	if dErr != nil {
		log.Fatalf("Failed to delete blog: %v\n", dErr)
	}
	// receives a response
	fmt.Printf("Blog: %v deleted\n", res.GetBlogId())
}

//doUpdateBlog sends a request for a updated blog
func doUpdateBlog(c blogpb.BlogServiceClient) {
	fmt.Println("Updating a Blog")
	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       "5eb9309c7ade23cccb8c797f",
			AuthorId: "Ronald",
			Content:  "This is to test my updated blog",
			Title:    "I Updated My Blog",
		},
	}
	res, rErr := c.UpdateBlog(context.Background(), req)
	if rErr != nil {
		log.Fatalf("A server error has accured: %v\n", rErr)
		return
	}
	fmt.Printf("blog %v updated\n", res.GetBlog())
}

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
	// doUnary(c)

	// doReadBlog(c)

	// doUpdateBlog(c)

	// doDeleteBlog() Deletes a blog
	doDeleteBlog(c)

}
