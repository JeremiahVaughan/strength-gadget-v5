import { render } from '@testing-library/react';

import ErrorNotification from './error-notification';

describe('ErrorNotification', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ErrorNotification />);
    expect(baseElement).toBeTruthy();
  });
});
