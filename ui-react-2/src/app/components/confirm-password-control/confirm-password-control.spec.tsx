import { render } from '@testing-library/react';

import ConfirmPasswordControl from './confirm-password-control';

describe('ConfirmPasswordControl', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ConfirmPasswordControl />);
    expect(baseElement).toBeTruthy();
  });
});
