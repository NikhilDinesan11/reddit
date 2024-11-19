package main

import (
	"fmt"
	"sync"
	"time"
)

// Data structures remain the same
type User struct {
	Username string
	Karma    int
	JoinedAt time.Time
}

type Post struct {
	ID        int
	Author    string
	Content   string
	Score     int
	Comments  []*Comment
	CreatedAt time.Time
}

type Comment struct {
	ID        int
	Author    string
	Content   string
	Score     int
	Replies   []*Comment
	CreatedAt time.Time
}

type Subreddit struct {
	Name        string
	Description string
	Members     map[string]bool
	Posts       []*Post
	CreatedAt   time.Time
}

type DirectMessage struct {
	ID        int
	From      string
	To        string
	Content   string
	CreatedAt time.Time
}

// Message interface
type Message interface {
	MessageType() string
}

// User Management Messages
type RegisterUserMsg struct {
	Username string
	Reply    chan error
}

func (m RegisterUserMsg) MessageType() string {
	return "RegisterUser"
}

// Subreddit Messages
type CreateSubredditMsg struct {
	Name        string
	Description string
	Creator     string
	Reply       chan error
}

func (m CreateSubredditMsg) MessageType() string {
	return "CreateSubreddit"
}

type JoinSubredditMsg struct {
	Username      string
	SubredditName string
	Reply         chan error
}

func (m JoinSubredditMsg) MessageType() string {
	return "JoinSubreddit"
}

type LeaveSubredditMsg struct {
	Username      string
	SubredditName string
	Reply         chan error
}

func (m LeaveSubredditMsg) MessageType() string {
	return "LeaveSubreddit"
}

// Post Messages
type CreatePostMsg struct {
	SubredditName string
	Author        string
	Content       string
	Reply         chan error
}

func (m CreatePostMsg) MessageType() string {
	return "CreatePost"
}

type CreateCommentMsg struct {
	PostID    int
	ParentID  int
	Author    string
	Content   string
	Reply     chan error
}

func (m CreateCommentMsg) MessageType() string {
	return "CreateComment"
}

type VoteMsg struct {
	ItemType string // "post" or "comment"
	ItemID   int
	Vote     int // 1 for upvote, -1 for downvote
	Reply    chan error
}

func (m VoteMsg) MessageType() string {
	return "Vote"
}

// Direct Message
type SendDMMsg struct {
	From    string
	To      string
	Content string
	Reply   chan error
}

func (m SendDMMsg) MessageType() string {
	return "SendDM"
}

// Actors
type RedditEngine struct {
	users          map[string]*User
	subreddits     map[string]*Subreddit
	directMessages map[string][]*DirectMessage
	nextPostID     int
	nextCommentID  int
	nextDMID       int
	mu             sync.RWMutex
	
	// Channels for different message types
	userChan      chan Message
	subredditChan chan Message
	postChan      chan Message
	dmChan        chan Message
}

func NewRedditEngine() *RedditEngine {
	engine := &RedditEngine{
		users:          make(map[string]*User),
		subreddits:     make(map[string]*Subreddit),
		directMessages: make(map[string][]*DirectMessage),
		userChan:       make(chan Message, 100),
		subredditChan:  make(chan Message, 100),
		postChan:       make(chan Message, 100),
		dmChan:         make(chan Message, 100),
	}
	
	// Start actor routines
	go engine.userActor()
	go engine.subredditActor()
	go engine.postActor()
	go engine.dmActor()
	
	return engine
}

func (re *RedditEngine) userActor() {
	for msg := range re.userChan {
		switch m := msg.(type) {
		case RegisterUserMsg:
			re.mu.Lock()
			if _, exists := re.users[m.Username]; exists {
				re.mu.Unlock()
				m.Reply <- fmt.Errorf("username already taken")
				continue
			}
			
			re.users[m.Username] = &User{
				Username: m.Username,
				JoinedAt: time.Now(),
			}
			re.mu.Unlock()
			m.Reply <- nil
		}
	}
}

func (re *RedditEngine) subredditActor() {
	for msg := range re.subredditChan {
		switch m := msg.(type) {
		case CreateSubredditMsg:
			re.mu.Lock()
			if _, exists := re.subreddits[m.Name]; exists {
				re.mu.Unlock()
				m.Reply <- fmt.Errorf("subreddit already exists")
				continue
			}
			
			re.subreddits[m.Name] = &Subreddit{
				Name:        m.Name,
				Description: m.Description,
				Members:     make(map[string]bool),
				CreatedAt:   time.Now(),
			}
			re.mu.Unlock()
			m.Reply <- nil
			
		case JoinSubredditMsg:
			re.mu.Lock()
			if subreddit, exists := re.subreddits[m.SubredditName]; exists {
				subreddit.Members[m.Username] = true
				re.mu.Unlock()
				m.Reply <- nil
			} else {
				re.mu.Unlock()
				m.Reply <- fmt.Errorf("subreddit not found")
			}
			
		case LeaveSubredditMsg:
			re.mu.Lock()
			if subreddit, exists := re.subreddits[m.SubredditName]; exists {
				delete(subreddit.Members, m.Username)
				re.mu.Unlock()
				m.Reply <- nil
			} else {
				re.mu.Unlock()
				m.Reply <- fmt.Errorf("subreddit not found")
			}
		}
	}
}

func (re *RedditEngine) postActor() {
	for msg := range re.postChan {
		switch m := msg.(type) {
		case CreatePostMsg:
			re.mu.Lock()
			if subreddit, exists := re.subreddits[m.SubredditName]; exists {
				post := &Post{
					ID:        re.nextPostID,
					Author:    m.Author,
					Content:   m.Content,
					CreatedAt: time.Now(),
				}
				re.nextPostID++
				subreddit.Posts = append(subreddit.Posts, post)
				re.mu.Unlock()
				m.Reply <- nil
			} else {
				re.mu.Unlock()
				m.Reply <- fmt.Errorf("subreddit not found")
			}
			
		case CreateCommentMsg:
			re.mu.Lock()
			comment := &Comment{
				ID:        re.nextCommentID,
				Author:    m.Author,
				Content:   m.Content,
				CreatedAt: time.Now(),
			}
			re.nextCommentID++
			
			found := false
			for _, subreddit := range re.subreddits {
				for _, post := range subreddit.Posts {
					if post.ID == m.PostID {
						if m.ParentID == 0 {
							post.Comments = append(post.Comments, comment)
							found = true
						} else {
							found = re.addReplyToComment(post.Comments, m.ParentID, comment)
						}
						break
					}
				}
				if found {
					break
				}
			}
			re.mu.Unlock()
			
			if !found {
				m.Reply <- fmt.Errorf("post or parent comment not found")
			} else {
				m.Reply <- nil
			}
			
		case VoteMsg:
			re.mu.Lock()
			if m.ItemType == "post" {
				found := false
				for _, subreddit := range re.subreddits {
					for _, post := range subreddit.Posts {
						if post.ID == m.ItemID {
							post.Score += m.Vote
							re.users[post.Author].Karma += m.Vote
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				re.mu.Unlock()
				if !found {
					m.Reply <- fmt.Errorf("post not found")
				} else {
					m.Reply <- nil
				}
			} else {
				// Handle comment votes similarly
				re.mu.Unlock()
				m.Reply <- nil
			}
		}
	}
}

func (re *RedditEngine) dmActor() {
	for msg := range re.dmChan {
		switch m := msg.(type) {
		case SendDMMsg:
			re.mu.Lock()
			dm := &DirectMessage{
				ID:        re.nextDMID,
				From:      m.From,
				To:        m.To,
				Content:   m.Content,
				CreatedAt: time.Now(),
			}
			re.nextDMID++
			
			// Store DM for both sender and receiver
			re.directMessages[m.From] = append(re.directMessages[m.From], dm)
			re.directMessages[m.To] = append(re.directMessages[m.To], dm)
			re.mu.Unlock()
			m.Reply <- nil
		}
	}
}

// Helper function to recursively add replies to comments
func (re *RedditEngine) addReplyToComment(comments []*Comment, parentID int, newComment *Comment) bool {
	for _, comment := range comments {
		if comment.ID == parentID {
			comment.Replies = append(comment.Replies, newComment)
			return true
		}
		if found := re.addReplyToComment(comment.Replies, parentID, newComment); found {
			return true
		}
	}
	return false
}

// Public API methods remain the same
func (re *RedditEngine) RegisterUser(username string) error {
	reply := make(chan error)
	re.userChan <- RegisterUserMsg{Username: username, Reply: reply}
	return <-reply
}

func (re *RedditEngine) CreateSubreddit(name, description, creator string) error {
	reply := make(chan error)
	re.subredditChan <- CreateSubredditMsg{Name: name, Description: description, Creator: creator, Reply: reply}
	return <-reply
}

func (re *RedditEngine) JoinSubreddit(username, subredditName string) error {
	reply := make(chan error)
	re.subredditChan <- JoinSubredditMsg{Username: username, SubredditName: subredditName, Reply: reply}
	return <-reply
}

func (re *RedditEngine) CreatePost(subredditName, author, content string) error {
	reply := make(chan error)
	re.postChan <- CreatePostMsg{SubredditName: subredditName, Author: author, Content: content, Reply: reply}
	return <-reply
}

func (re *RedditEngine) CreateComment(postID, parentID int, author, content string) error {
	reply := make(chan error)
	re.postChan <- CreateCommentMsg{PostID: postID, ParentID: parentID, Author: author, Content: content, Reply: reply}
	return <-reply
}

func (re *RedditEngine) Vote(itemType string, itemID int, vote int) error {
	reply := make(chan error)
	re.postChan <- VoteMsg{ItemType: itemType, ItemID: itemID, Vote: vote, Reply: reply}
	return <-reply
}

func (re *RedditEngine) SendDM(from, to, content string) error {
	reply := make(chan error)
	re.dmChan <- SendDMMsg{From: from, To: to, Content: content, Reply: reply}
	return <-reply
}

func main() {
	engine := NewRedditEngine()

	// Register users
	engine.RegisterUser("alice")
	engine.RegisterUser("bob")

	// Create and join subreddit
	engine.CreateSubreddit("golang", "All about Go programming", "alice")
	engine.JoinSubreddit("bob", "golang")

	// Create post and comments
	engine.CreatePost("golang", "alice", "Actor models are awesome!")
	engine.CreateComment(0, 0, "bob", "Totally agree!")
	engine.CreateComment(0, 1, "alice", "Thanks!")

	// Vote and send DMs
	engine.Vote("post", 0, 1)
	engine.SendDM("alice", "bob", "Thanks for the support!")
}