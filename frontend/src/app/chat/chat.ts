import { CommonModule } from '@angular/common';
import { Component, OnInit, OnDestroy } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { Subscription } from 'rxjs';
import { ChatMessage, ChatService } from '../service/chat';

@Component({
  selector: 'app-chat',
    standalone: true,
    imports: [FormsModule, CommonModule],
  templateUrl: './chat.html',
  styleUrls: ['./chat.scss']
})
export class Chat implements OnInit, OnDestroy {
  messages: ChatMessage[] = [];
  messageInput: string = '';
  username: string = '';
  room: string = 'general';
  isConnected: boolean = false;
  showUsernameInput: boolean = true;
  
  private messagesSubscription?: Subscription;
  private connectionSubscription?: Subscription;

  constructor(private chatService: ChatService) {}

  ngOnInit(): void {
    // Subscribe to messages
    this.messagesSubscription = this.chatService.messages$.subscribe((message) => {
      this.messages.push(message);
      // Auto-scroll to bottom
      setTimeout(() => this.scrollToBottom(), 100);
    });

    // Subscribe to connection status
    this.connectionSubscription = this.chatService.connectionStatus$.subscribe((status) => {
      this.isConnected = status;
    });
  }

  ngOnDestroy(): void {
    this.chatService.disconnect();
    this.messagesSubscription?.unsubscribe();
    this.connectionSubscription?.unsubscribe();
  }

  joinChat(): void {
    if (this.username.trim()) {
      this.chatService.connect(this.username, this.room);
      this.showUsernameInput = false;
    }
  }

  sendMessage(): void {
    if (this.messageInput.trim() && this.isConnected) {
      this.chatService.sendMessage(this.messageInput);
      this.messageInput = '';
    }
  }

  leaveChat(): void {
    this.chatService.disconnect();
    this.showUsernameInput = true;
    this.messages = [];
  }

  private scrollToBottom(): void {
    const messagesContainer = document.querySelector('.messages-container');
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  }
}