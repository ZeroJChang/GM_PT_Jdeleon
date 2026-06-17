import { Controller, Get, Param, Post, UseGuards } from '@nestjs/common';

import { BooksService } from '../../books/books.service';
import { InternalApiKeyGuard } from '../guards/internal-api-key.guard';

@Controller('internal/books')
@UseGuards(InternalApiKeyGuard)
export class InternalBooksController {
  constructor(private readonly booksService: BooksService) {}

  @Get(':id')
  getBook(@Param('id') id: string) {
    return this.booksService.getInternalBook(id);
  }

  @Post(':id/reserve')
  reserveBook(@Param('id') id: string) {
    return this.booksService.reserveCopy(id);
  }

  @Post(':id/release')
  releaseBook(@Param('id') id: string) {
    return this.booksService.releaseCopy(id);
  }
}
