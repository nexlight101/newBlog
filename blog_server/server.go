package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/nexlight101/gRPC_course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
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

// DeleteBlog Deletes a blog
func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Deleting blog")
	// receive a requset
	blogID := req.GetBlogId()
	// convert ID to primitive object id
	bID, pErr := primitive.ObjectIDFromHex(blogID)
	if pErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert object ID: %v\n", pErr))
	}
	// Delete the blog in mongoDB
	filter := bson.M{
		"_id": bID,
	}
	result, dErr := collection.DeleteOne(context.Background(), filter)
	if dErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Delete opperation failed %v\n", dErr))
	}
	if result.DeletedCount == 0 {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Blog was not deleted: %v\n", blogID))
	}
	// send back a response
	return &blogpb.DeleteBlogResponse{
		BlogId: blogID,
	}, nil

}

// UpdateBlog updates a blog
func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("UpdateBlog request received")
	// receive a request
	blog := req.GetBlog()
	// find and update a blog
	blogID := blog.GetId()
	bID, cErr := primitive.ObjectIDFromHex(blogID)
	if cErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert id %v\n", cErr))
	}
	filter := bson.M{"_id": bID}
	update := bson.M{"$set": blogItem{
		AuthorID: blog.AuthorId,
		Content:  blog.Content,
		Title:    blog.Title,
	},
	}
	fmt.Println("Updating record in mongoDB")
	result, uErr := collection.UpdateOne(context.Background(), filter, update)
	if uErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot write record %v\n", uErr))
	}
	if result.MatchedCount == 0 {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot find record: %v\n", blogID))
	}
	// responds with an updated blog
	res := &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       blogID,
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
			Title:    blog.GetTitle(),
		},
	}
	return res, nil
}

// ReadBlog reads a specific blog from blogID
func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request received")
	// Convert the request object id to primitive object id
	blogID, pErr := primitive.ObjectIDFromHex(req.GetBlogId())
	if pErr != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not convert object id: %v\n", pErr))
	}

	// Find the blog in mongo
	filter := bson.M{"_id": blogID}
	blog := collection.FindOne(context.Background(), filter)
	if blog.Err() != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Could not find object id: %v\n", blog.Err()))
	}
	// Decode the blog
	foundBlog := blogItem{}
	dErr := blog.Decode(&foundBlog)
	if dErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Could decode blog: %v\n", dErr))
	}
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       req.GetBlogId(),
			AuthorId: foundBlog.AuthorID,
			Content:  foundBlog.Content,
			Title:    foundBlog.Title,
		},
	}, nil
}

//Creates a blog item and save it in mongo db
func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Blog request received")
	blog := req.GetBlog()
	// Make a blog
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}
	// Store the blog
	res, mErr := collection.InsertOne(context.Background(), data)
	// Handle mongo Errors
	if mErr != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal mongoDB Error, %v\n", mErr))
	}

	// return the response (blog)
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to oid, %v\n", mErr))
	}
	response := blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}

	return &response, nil
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
