import { BadRequestException, NotFoundException } from '@nestjs/common';

jest.mock('../database/prisma.service', () => ({
  PrismaService: class PrismaService {},
}));

import { BooksService } from './books.service';

describe('BooksService', () => {
  let service: BooksService;

  const prisma = {
    book: {
      create: jest.fn(),
      findMany: jest.fn(),
      count: jest.fn(),
      findUnique: jest.fn(),
      updateMany: jest.fn(),
      update: jest.fn(),
      delete: jest.fn(),
    },
  };

  const book = {
    id: '117f2aaf-61ab-4b49-b34a-7b331f6947a8',
    title: 'Clean Code',
    author: 'Robert C. Martin',
    isbn: '9780132350884',
    year: 2008,
    genre: 'Software Engineering',
    totalCopies: 5,
    availableCopies: 5,
    createdAt: new Date('2026-01-01T00:00:00.000Z'),
    updatedAt: new Date('2026-01-01T00:00:00.000Z'),
  };

  beforeEach(() => {
    jest.clearAllMocks();
    service = new BooksService(prisma as never);
  });

  it('should create a book with availableCopies equal to totalCopies', async () => {
    const createBookDto = {
      title: 'Clean Code',
      author: 'Robert C. Martin',
      isbn: '9780132350884',
      year: 2008,
      genre: 'Software Engineering',
      totalCopies: 5,
    };

    prisma.book.create.mockResolvedValue(book);

    const result = await service.create(createBookDto);

    expect(prisma.book.create).toHaveBeenCalledWith({
      data: {
        ...createBookDto,
        availableCopies: 5,
      },
    });

    expect(result).toEqual(book);
  });

  it('should return paginated books using filters', async () => {
    prisma.book.findMany.mockResolvedValue([book]);
    prisma.book.count.mockResolvedValue(1);

    const result = await service.findAll({
      author: 'Robert',
      available: true,
      page: 1,
      limit: 10,
    });

    expect(prisma.book.findMany).toHaveBeenCalledWith(
      expect.objectContaining({
        skip: 0,
        take: 10,
        orderBy: {
          createdAt: 'desc',
        },
        where: expect.objectContaining({
          author: {
            contains: 'Robert',
            mode: 'insensitive',
          },
          availableCopies: {
            gt: 0,
          },
        }),
      }),
    );

    expect(result).toEqual({
      items: [book],
      meta: {
        total: 1,
        page: 1,
        limit: 10,
        totalPages: 1,
      },
    });
  });

  it('should throw NotFoundException when book does not exist', async () => {
    prisma.book.findUnique.mockResolvedValue(null);

    await expect(service.findOne('missing-book-id')).rejects.toThrow(
      NotFoundException,
    );
  });

  it('should reserve one available copy', async () => {
    const reservedBook = {
      ...book,
      availableCopies: 4,
    };

    prisma.book.findUnique
      .mockResolvedValueOnce(book)
      .mockResolvedValueOnce(reservedBook);

    prisma.book.updateMany.mockResolvedValue({
      count: 1,
    });

    const result = await service.reserveCopy(book.id);

    expect(prisma.book.updateMany).toHaveBeenCalledWith({
      where: {
        id: book.id,
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

    expect(result).toEqual({
      id: reservedBook.id,
      title: reservedBook.title,
      author: reservedBook.author,
      isbn: reservedBook.isbn,
      year: reservedBook.year,
      genre: reservedBook.genre,
      totalCopies: reservedBook.totalCopies,
      availableCopies: reservedBook.availableCopies,
      isAvailable: true,
    });
  });

  it('should throw BadRequestException when there are no available copies', async () => {
    const unavailableBook = {
      ...book,
      availableCopies: 0,
    };

    prisma.book.findUnique.mockResolvedValue(unavailableBook);
    prisma.book.updateMany.mockResolvedValue({
      count: 0,
    });

    await expect(service.reserveCopy(book.id)).rejects.toThrow(
      BadRequestException,
    );
  });
});
