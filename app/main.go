package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"github.com/gorilla/mux"
)

type BookCheckout struct{
	BookID string 	`json:"book_id"`
	User   string 	`json:"user"`
	CheckoutDate 	string 	`json:"checkout_date"`
	IsGenesis bool `json:"is_genesis"`
}

type Book struct{
	ID string 	`json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN string `json:"isbn"`
}

type Block struct{
	Hash string
	PrevHash string
	Data BookCheckout
	Position int
	Timestamp string
}


type Blockchain struct{
	blocks []*Block
}

var BlockChain *Blockchain;

func (b *Block)generateHash(){
	bytes,_ := json.Marshal(b.Data)
	data := string(b.Position)+b.Timestamp+string(bytes)+b.PrevHash;

	hash := sha256.New();
	hash.Write([]byte(data));
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func CreateBlock(prevBlock *Block,checkoutItem BookCheckout) *Block{
	block := &Block{}
	block.Position = prevBlock.Position+1;
	block.Timestamp = time.Now().String()
	block.Data = checkoutItem;
	block.PrevHash = prevBlock.Hash;
	block.generateHash();

	return block;
}

func (block *Block) validateHash(hash string)bool{
	block.generateHash();
	if block.Hash!= hash{
		return false;
	}
	return true;
}

func validBlock(block *Block,prevBlock *Block) bool{
	if prevBlock.Hash!=block.PrevHash{
		return false;
	}
	if !block.validateHash(block.Hash){
		return false;
	}
	if prevBlock.Position+1!=block.Position{
		return false;
	}
	return true;
}

func (bc *Blockchain)AddBlock(data BookCheckout){
	prevBlock := bc.blocks[len(bc.blocks)-1];
	block := CreateBlock(prevBlock,data)

	if validBlock(block,prevBlock){
		bc.blocks = append(bc.blocks, block)
	}
}

func GenesisBlock() *Block{
	return CreateBlock(&Block{},BookCheckout{IsGenesis: true});
}

func NewBlockchain() *Blockchain{
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func getBlockchain(w http.ResponseWriter,r *http.Request){
	bytes,err := json.MarshalIndent(BlockChain.blocks,"","")

	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		json.NewEncoder(w).Encode(err)
		return;
	}

	io.WriteString(w,string(bytes));
}

func WriteBlock(w http.ResponseWriter,r *http.Request){
	var checkoutItem BookCheckout

	if err:=json.NewDecoder(r.Body).Decode(&checkoutItem);err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		log.Println("Could not write Block:",err)
		w.Write([]byte("Could not write block"))
		return;
	}

	BlockChain.AddBlock(checkoutItem)

	resp,err := json.MarshalIndent(checkoutItem,"","");

	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not marshall paylod %v",err)
		w.Write([]byte("could not write Block"))
		return;
	}
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}


func NewBook(w http.ResponseWriter,r *http.Request){
	var book Book
	if err:= json.NewDecoder(r.Body).Decode(&book);err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		log.Println("Could not create ",err)
		w.Write([]byte("Could not create new Book"))
		return;
	}

	h := md5.New();
	io.WriteString(h,book.ISBN+book.PublishDate)
	book.ID = fmt.Sprintf("%x",h.Sum(nil))

	resp,err := json.MarshalIndent(book,"","")

	if err!=nil{
		w.WriteHeader(http.StatusInternalServerError);
		log.Printf("Could not marshal payload, %v",err)
		w.Write([]byte("Could not save book data"));return;
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp);
}

func main(){
	BlockChain = NewBlockchain()

	r := mux.NewRouter();

	r.HandleFunc("/",getBlockchain).Methods("GET");
	r.HandleFunc("/",WriteBlock).Methods("POST");
	r.HandleFunc("/new",NewBook).Methods("POST");

	go func(){
		for _,block:=range BlockChain.blocks{
			fmt.Printf("Prev.Hash %x\n",block.PrevHash)
			bytes,_ := json.MarshalIndent(block.Data,"","")
			fmt.Printf("Data: %v\n",string(bytes))
			fmt.Printf("Hash: %v\n",string(block.Hash))
			fmt.Println();
		}
	}()
	fmt.Println("Hello")


	log.Println("Listening to port 3000");
	log.Fatal(http.ListenAndServe(":3000",r));
}