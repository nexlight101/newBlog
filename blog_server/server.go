package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nexlight101/gRPC_course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Define  your globle server struct
type server struct{}

// Define mongodb collection
var collection *mongo.Collection

// Define mongodb struct(table)
type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func main() {
	// if we crash the go code, we get the filename and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	//Define your mongo client
	fmt.Println("Connecting to MongoDB")
	app := "mongodb://localhost:27017"
	mOpts := []*options.ClientOptions{
		{
			AppName: &app,
		},
	}
	client, err := mongo.NewClient(mOpts...)
	if err != nil {
		log.Fatalf("Cannot create mongo client: %v\n", err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatalf("Cannot connect to Mongodb: %v\n", err)
	}

	// Connect to Mongodb collection(table)
	collection = client.Database("mydb").Collection("blog")
	// Create listener
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	fmt.Println("Blog Service Started")

	//Create a new gRPC server
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})
	go func() {
		fmt.Println("Starting gRPC Server...")
		// Check if the server is serving the listener
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	// Register reflection service on server
	reflection.Register(s)
	// wait for control c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	fmt.Println("End of Program")

}
