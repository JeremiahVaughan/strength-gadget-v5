import { render } from '@testing-library/react';

import ForgotPasswordNewPasswordPage from './forgot-password-new-password-page';

describe('ForgotPasswordNewPasswordPage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ForgotPasswordNewPasswordPage />);
    expect(baseElement).toBeTruthy();
  });
});
