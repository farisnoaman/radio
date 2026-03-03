import { Pagination, PaginationProps, useListContext } from 'react-admin';
import { useEffect } from 'react';

const DEFAULT_PAGE_SIZES = [10, 50, 100, 200, 500, 1000];

const ServerPagination = (props: PaginationProps) => {
  const { perPage, setPerPage } = useListContext();
  
  useEffect(() => {
    if (perPage > 0 && !DEFAULT_PAGE_SIZES.includes(perPage)) {
      setPerPage(DEFAULT_PAGE_SIZES[0]);
    }
  }, [perPage, setPerPage]);

  const rowsPerPageOptions = DEFAULT_PAGE_SIZES.includes(perPage) 
    ? DEFAULT_PAGE_SIZES 
    : [...new Set([...DEFAULT_PAGE_SIZES, perPage])].sort((a, b) => a - b);

  return <Pagination rowsPerPageOptions={rowsPerPageOptions} {...props} />;
};

export { ServerPagination, DEFAULT_PAGE_SIZES };
