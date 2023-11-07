import { render } from '@testing-library/react';

import ForgotPasswordEmailPage from './forgot-password-email-page';

describe('ForgotEmailPage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ForgotPasswordEmailPage />);
    expect(baseElement).toBeTruthy();
  });
});
