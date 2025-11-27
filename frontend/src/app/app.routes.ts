import { Routes } from '@angular/router';
import { Chat } from './chat/chat';

export const routes: Routes = [
    {
    path: 'chat',
    component: Chat,
  },
  {
    path: '',
    redirectTo: 'chat',
    pathMatch: 'full'
  }
];
