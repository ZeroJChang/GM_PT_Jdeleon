import {
  BadRequestException,
  Injectable,
  NotFoundException,
} from '@nestjs/common';

import { PrismaService } from '../database/prisma.service';
import { CreateBookDto } from './dto/create-book.dto';
import { QueryBooksDto } from './dto/query-books.dto';
import { UpdateBookDto } from './dto/update-book.dto';

@Injectable()
export class BooksService {
  constructor(private readonly prisma: PrismaService) {}

  async create(createBookDto: CreateBookDto) {
    return this.prisma.book.create({
      data: {
        ...createBookDto,
        availableCopies: createBookDto.totalCopies,
      },
    });
  }

  async findAll(query: QueryBooksDto) {
    const page = query.page ?? 1;
    const limit = query.limit ?? 10;
    const skip = (page - 1) * limit;

    const where = {
      ...(query.author && {
        author: {
          contains: query.author,
          mode: 'insensitive' as const,
        },
      }),
      ...(query.genre && {
        genre: {
          contains: query.genre,
          mode: 'insensitive' as const,
        },
      }),
      ...(query.available !== undefined && {
        availableCopies: query.available ? { gt: 0 } : { equals: 0 },
      }),
    };

    const [items, total] = await Promise.all([
      this.prisma.book.findMany({
        where,
        skip,
        take: limit,
        orderBy: {
          createdAt: 'desc',
        },
      }),
      this.prisma.book.count({ where }),
    ]);

    return {
      items,
      meta: {
        total,
        page,
        limit,
        totalPages: Math.ceil(total / limit),
      },
    };
  }

  async findOne(id: string) {
    const book = await this.prisma.book.findUnique({
      where: { id },
    });

    if (!book) {
      throw new NotFoundException('Book not found');
    }

    return book;
  }

  async update(id: string, updateBookDto: UpdateBookDto) {
    const currentBook = await this.findOne(id);

    const borrowedCopies =
      currentBook.totalCopies - currentBook.availableCopies;

    if (
      updateBookDto.totalCopies !== undefined &&
      updateBookDto.totalCopies < borrowedCopies
    ) {
      throw new BadRequestException(
        'Total copies cannot be lower than borrowed copies',
      );
    }

    return this.prisma.book.update({
      where: { id },
      data: {
        ...updateBookDto,
        ...(updateBookDto.totalCopies !== undefined && {
          availableCopies: updateBookDto.totalCopies - borrowedCopies,
        }),
      },
    });
  }

  async remove(id: string) {
    await this.findOne(id);

    await this.prisma.book.delete({
      where: { id },
    });

    return {
      message: 'Book deleted successfully',
    };
  }

  async getInternalBook(id: string) {
    const book = await this.findOne(id);

    return {
      id: book.id,
      title: book.title,
      author: book.author,
      isbn: book.isbn,
      year: book.year,
      genre: book.genre,
      totalCopies: book.totalCopies,
      availableCopies: book.availableCopies,
      isAvailable: book.availableCopies > 0,
    };
  }

  async reserveCopy(id: string) {
    await this.findOne(id);

    const result = await this.prisma.book.updateMany({
      where: {
        id,
        availableCopies: {
          gt: 0,
        },
      },
      data: {
        availableCopies: {
          decrement: 1,
        },
      },
    });

    if (result.count === 0) {
      throw new BadRequestException('Book has no available copies');
    }

    return this.getInternalBook(id);
  }

  async releaseCopy(id: string) {
    const book = await this.findOne(id);

    if (book.availableCopies >= book.totalCopies) {
      throw new BadRequestException('Book already has all copies available');
    }

    await this.prisma.book.update({
      where: { id },
      data: {
        availableCopies: {
          increment: 1,
        },
      },
    });

    return this.getInternalBook(id);
  }
}