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

    if (
      updateBookDto.totalCopies !== undefined &&
      updateBookDto.totalCopies < currentBook.totalCopies - currentBook.availableCopies
    ) {
      throw new BadRequestException(
        'Total copies cannot be lower than borrowed copies',
      );
    }

    const borrowedCopies = currentBook.totalCopies - currentBook.availableCopies;

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
}