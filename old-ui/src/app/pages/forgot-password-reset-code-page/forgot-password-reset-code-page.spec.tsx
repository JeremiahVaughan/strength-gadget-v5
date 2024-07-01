import { render } from '@testing-library/react';

import ForgotPasswordResetCodePage from './forgot-password-reset-code-page';

describe('ForgotPasswordResetCodePage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ForgotPasswordResetCodePage />);
    expect(baseElement).toBeTruthy();
  });
});
