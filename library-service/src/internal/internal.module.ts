import { Module } from '@nestjs/common';

import { BooksModule } from '../books/books.module';
import { InternalBooksController } from './controllers/internal-books.controller';

@Module({
  imports: [BooksModule],
  controllers: [InternalBooksController],
})
export class InternalModule {}
