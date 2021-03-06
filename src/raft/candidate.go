package raft

func (rf *Raft) election() {
	// fmt.Println("[electing]", rf.me, " - ", rf.currentTerm)

	lastEntry := rf.lastLogEntry()
	currentTerm := rf.currentTerm
	args := &RequestVoteArgs{
		Term:         currentTerm,
		CandidateID:  rf.me,
		LastLogIndex: lastEntry.Index,
		LastLogTerm:  lastEntry.Term,
	}

	rf.newVotedFor(rf.me)
	votes := 1

	for peer := range rf.peers {
		if peer == rf.me {
			continue
		}

		go func(peer int) {
			reply := &RequestVoteReply{}
			// fmt.Println("requesting vote to: ", peer, " req: ", args)
			ok := rf.sendRequestVote(peer, args, reply)
			// fmt.Println("reply vote from: ", peer, " reply: ", reply)
			if !ok {
				return
			}

			rf.mu.Lock()
			defer rf.mu.Unlock()

			if rf.currentTerm != currentTerm || rf.state != Candidate {
				return
			}

			if reply.VoteGranted {
				votes++
				if votes >= len(rf.peers)/2+1 {
					rf.switchState(Leader, rf.currentTerm)
					rf.broadcastHeartbeats()
				}
			} else if reply.Term > rf.currentTerm {
				rf.switchState(Follower, reply.Term)
			}
		}(peer)
	}
}
