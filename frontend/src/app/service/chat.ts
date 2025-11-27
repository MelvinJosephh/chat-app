import { Injectable } from '@angular/core';
import { Subject, Observable } from 'rxjs';

export interface ChatMessage {
  type: string;
  username: string;
  content: string;
  room: string;
  timestamp: string;
}

@Injectable({
  providedIn: 'root'
})
export class ChatService {
  private socket: WebSocket | null = null;
  private messagesSubject = new Subject<ChatMessage>();
  private connectionStatusSubject = new Subject<boolean>();
  
  public messages$: Observable<ChatMessage> = this.messagesSubject.asObservable();
  public connectionStatus$: Observable<boolean> = this.connectionStatusSubject.asObservable();
  
  private isConnected = false;

  constructor() { }

  connect(username: string, room: string = 'general'): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      console.log('Already connected');
      return;
    }

    const url = `ws://localhost:8080/ws?username=${encodeURIComponent(username)}&room=${encodeURIComponent(room)}`;
    
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      console.log('✅ Connected to chat server');
      this.isConnected = true;
      this.connectionStatusSubject.next(true);
    };

    this.socket.onmessage = (event) => {
      const message: ChatMessage = JSON.parse(event.data);
      this.messagesSubject.next(message);
    };

    this.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      this.isConnected = false;
      this.connectionStatusSubject.next(false);
    };

    this.socket.onclose = () => {
      console.log('🔌 Disconnected from chat server');
      this.isConnected = false;
      this.connectionStatusSubject.next(false);
    };
  }

  sendMessage(content: string): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      const message = {
        type: 'message',
        content: content
      };
      this.socket.send(JSON.stringify(message));
    } else {
      console.error('WebSocket is not connected');
    }
  }

  disconnect(): void {
    if (this.socket) {
      this.socket.close();
      this.socket = null;
      this.isConnected = false;
      this.connectionStatusSubject.next(false);
    }
  }

  getConnectionStatus(): boolean {
    return this.isConnected;
  }
}